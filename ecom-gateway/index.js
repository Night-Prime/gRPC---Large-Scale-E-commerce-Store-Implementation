const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const { rateLimitAndTimeout } = require("./middleware");
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");
const path = require("path");

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

// Initialize gRPC Clients
const orderClient = loadProto("order/v1/order.proto", "order.v1", "OrderService", 8080);
const userClient = loadProto("user/v1/user.proto", "user.v1", "UserService", 8081);
const paymentClient = loadProto("payment/v1/payment.proto", "payment.v1", "PaymentService", 8082);

// App Instance
const app = express();

app.use(express.json());
app.use(cors());
app.use(helmet());
app.use(morgan("combined"));
app.disable("x-powered-by");

// Handle timeout & rate-limiting
app.use(rateLimitAndTimeout);

// ---------------------------------------------------------
// Expose HTTP Routes dynamically mapped to gRPC Services
// ---------------------------------------------------------
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
        order_status: ORDER_STATUS.UNSPECIFIED,
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
