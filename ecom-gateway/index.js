const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const { rateLimitAndTimeout } = require("./middleware");
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const path = require("path");
const { rateLimit } = require("express-rate-limit");
const client = require("prom-client");
const { METRIC_LABEL_ENUM, MetricsLabelClass, UsersMetricsLabelClass } = require("./metrics");

const PORT = process.env.PORT || 8000;

// Helper to Load gRPC Protos
const loadProto = (filename, packagePath, serviceName, port) => {
    const PROTO_PATH = path.join(__dirname, "../services/proto", filename);
    const packageDef = protoLoader.loadSync(PROTO_PATH, {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true
    });

    // Traverse the package path (e.g., "order.v1")
    let protoObj = grpc.loadPackageDefinition(packageDef);
    const parts = packagePath.split('.');
    for (const part of parts) {
        protoObj = protoObj[part];
    }

    return new protoObj[serviceName](`localhost:${port}`, grpc.credentials.createInsecure());
};

// Initialize gRPC Clients:
const orderClient = loadProto("order/v1/order.proto", "order.v1", "OrderService", 8080);
const userClient = loadProto("user/v1/user.proto", "user.v1", "UserService", 8081);
const paymentClient = loadProto("payment/v1/payment.proto", "payment.v1", "PaymentService", 8082);

// Rate Limiting setup:
const limiter = rateLimit({
    windowMs: 1 * 60 * 1000,
    max: 10,
    message: "Too many requests from this IP, please try again after 1 minute",
    standardHeaders: true,
    legacyHeaders: false,
});

// Metrics & Logging setup:
const register = new client.Registry();

// Getting the total request from an ip:
const http_request_total = new client.Counter({
    name: "http_request_total",
    help: "Total number of HTTP requests",
    labelNames: [METRIC_LABEL_ENUM.PATH, METRIC_LABEL_ENUM.METHOD, METRIC_LABEL_ENUM.STATUS_CODE, METRIC_LABEL_ENUM.SERVICE],
});

// Getting the gauge of the request:
const http_request_gauge = new client.Gauge({
    name: "http_request_gauge",
    help: "Gauge of HTTP requests",
    labelNames: [METRIC_LABEL_ENUM.PATH, METRIC_LABEL_ENUM.METHOD, METRIC_LABEL_ENUM.STATUS_CODE, METRIC_LABEL_ENUM.SERVICE, METRIC_LABEL_ENUM.USERS],
});

// Getting the response time:
const http_response_time = new client.Histogram({
    name: "http_response_time",
    help: "Response time of HTTP requests",
    labelNames: [METRIC_LABEL_ENUM.PATH, METRIC_LABEL_ENUM.METHOD, METRIC_LABEL_ENUM.STATUS_CODE, METRIC_LABEL_ENUM.SERVICE],
    buckets: [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
});

client.collectDefaultMetrics({
    register: register,
    prefix: "gateway_",
})

register.registerMetric(http_request_total);
register.registerMetric(http_request_gauge);
register.registerMetric(http_response_time);

// App Instance
const app = express();

// Middleware:
app.use(express.json());
app.use(cors());
app.use(helmet());
app.use(morgan("combined"));
app.disable("x-powered-by");
app.use(limiter);
app.use(rateLimitAndTimeout);

// Metrics & logging:
app.get("/metrics", async (req, res) => {
    res.set("Content-Type", register.contentType);
    res.end(await register.metrics());
});

// middleware to collect requests metrics:
app.use((req, res, next) => {
    const start = Date.now();

    // We record metrics when the response is finished
    res.on("finish", () => {
        const req_url = new URL(req.url, `http://${req.hostname}`);
        const pathname = req_url.pathname;
        const method = req.method;
        const statusCode = res.statusCode;
        const service = pathname.split("/")[1] || "gateway";
        const duration = (Date.now() - start) / 1000; // duration in seconds

        // Use the label names defined in the enum
        const labels = {
            [METRIC_LABEL_ENUM.PATH]: pathname,
            [METRIC_LABEL_ENUM.METHOD]: method,
            [METRIC_LABEL_ENUM.STATUS_CODE]: statusCode,
            [METRIC_LABEL_ENUM.SERVICE]: service
        };

        http_request_total.inc(labels);
        const gaugeLabels = {
            ...labels,
            [METRIC_LABEL_ENUM.USERS]: req.body.email || "anonymous"
        };
        http_request_gauge.inc(gaugeLabels);
        http_response_time.observe(labels, duration);
    });

    next();
});

// Order Status Enum:
const ORDER_STATUS = {
    UNSPECIFIED: 0,
    PENDING: 1,
    SUCCESS: 2,
    FAILED: 3,
}

// =============================================================
// User Service Routes
// =============================================================

app.post("/signup", (req, res) => {
    const payload = {
        email: req.body.email,
        password: req.body.password
    };

    userClient.Signup(payload, (error, response) => {
        if (error) return res.status(500).json({ code: 500, status: "Error", message: error.details || error.message });
        res.json({ code: 200, status: "Success", data: response });
    });
});


app.post("/login", (req, res) => {
    const payload = {
        email: req.body.email,
        password: req.body.password
    };

    userClient.Login(payload, (error, response) => {
        if (error) return res.status(500).json({ code: 500, status: "Error", message: error.details || error.message });
        res.json({ code: 200, status: "Success", data: response });
    });
});

// =============================================================
// Order Service Routes
// =============================================================

app.post("/order", (req, res) => {
    const payload = {
        order_id: req.body.order_id,
        order_product: req.body.order_product,
        is_paid: false,
        user_id: req.body.user_id,
        order_status: ORDER_STATUS.PENDING,
        order_amount: req.body.order_amount
    };

    orderClient.CreateOrder(payload, (error, response) => {
        if (error) return res.status(500).json({ code: 500, status: "Error", message: error.details || error.message });
        res.json({ code: 200, status: "Success", data: response });
    });
});

// =============================================================
// Payment Service Routes
// =============================================================

app.post("/payment", (req, res) => {
    const payload = {
        user_id: req.body.user_id,
        order_id: req.body.order_id,
        order_amount: req.body.order_amount
    };

    paymentClient.Charge(payload, (error, response) => {
        if (error) return res.status(500).json({ code: 500, status: "Error", message: error.details || error.message });
        res.json({ code: 200, status: "Success", data: response });
    });
});

app.get("/", (req, res) => {
    res.send("Welcome to E-com Gateway Service (Powered by pure gRPC backend)");
});

app.use((req, res) => {
    res.status(404).json({ code: 404, status: "Error", message: "Route not found", data: null });
});

app.listen(PORT, () => {
    console.log(`Running E-com Gateway on Port: ${PORT}`);
    console.log("____________________________________________");
});
