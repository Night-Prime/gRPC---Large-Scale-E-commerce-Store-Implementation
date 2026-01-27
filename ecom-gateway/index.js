const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const { rateLimitAndTimeout } = require("./middleware");
const { services } = require("./serviceList");
const { createProxyMiddleware } = require("http-proxy-middleware");
const PORT = process.env.PORT || 8000;


// app instance
const app = express();

app.use(cors());
app.use(helmet());
app.use(morgan("combined"));
app.disable("x-powered-by");

// handle timeout & rate-limiting
app.use(rateLimitAndTimeout);

// handling services routing:
services.forEach(({ route, target }) => {
    const proxyOptions = {
        target,
        changeOrigin: true,
    };
    console.log("Route: ", route, "options: ", proxyOptions);

    app.use(route, rateLimitAndTimeout, createProxyMiddleware(proxyOptions));
});

app.get("/", (req, res) => {
    res.send("Welcome to E-com Gateway Service")
})

app.use((req, res) => {
    res.status(404).json({
        code: 404,
        status: "Error",
        message: "Route not found",
        data: null
    })
})

app.listen(PORT, () => {
    console.log(`Running E-com Gateway on Port: ${PORT}`);
    console.log("____________________________________________")
})
