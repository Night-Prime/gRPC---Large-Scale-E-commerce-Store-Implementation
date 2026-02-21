// Handling my Middleware here for now:
const rateLimit = 10;
const interval = 60 * 1000;
let requestCount = {};

setInterval(() => {
    Object.keys(requestCount).forEach((ip) => {
        requestCount[ip] = 0;
    });
}, interval);

setTimeout(() => {
    requestCount = {};
}, interval);


export function rateLimitAndTimeout(req, res, next) {
    const ip = req.ip;

    requestCount[ip] = (requestCount[ip] || 0) + 1;
    if (requestCount[ip] > rateLimit) {
        return res.status(429).json({
            code: 429,
            status: 'Error',
            message: "Rate Limit Exceeded",
            data: null
        })
    }

    req.setTimeout(15000, () => {
        res.status(504).json({
            code: 504,
            status: "Error",
            message: "Gateway Timeout",
            data: null,
        })
    });

    next()
}

