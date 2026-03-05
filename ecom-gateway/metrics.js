const METRIC_LABEL_ENUM = {
    PATH: "pathname",
    METHOD: "method",
    STATUS_CODE: "statusCode",
    SERVICE: "service",
    USERS: "users",
}

class MetricsLabelClass {
    constructor(method, pathname, statusCode, service) {
        this.method = method
        this.pathname = pathname
        this.statusCode = statusCode
        this.service = service
    }
}

class UsersMetricsLabelClass {
    constructor(method, pathname, statusCode, service, users) {
        this.method = method
        this.pathname = pathname
        this.statusCode = statusCode
        this.service = service
        this.users = users
    }
}

module.exports = {
    METRIC_LABEL_ENUM,
    MetricsLabelClass,
    UsersMetricsLabelClass
};
