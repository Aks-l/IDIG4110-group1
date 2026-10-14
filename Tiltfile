docker_compose("./docker-compose.yml")

dc_resource(
    'api-gateway',
    labels=['api']
)

dc_resource(
    'ingest-service',
    labels=['ingest']
)

dc_resource(
    'rules-engine',
    labels=['rules']
)

dc_resource(
    'device-simulator',
    labels=['simulator']
)

dc_resource(
    'mosquitto',
    labels=['infra']
)

dc_resource(
    'kafka',
    labels=['infra']
)

dc_resource(
    'ingest-db',
    labels=['db']
)

dc_resource(
    'twin-db',
    labels=['db']
)

dc_resource(
    'rules-db',
    labels=['db']
)

dc_resource(
    'identity-db',
    labels=['db']
)
