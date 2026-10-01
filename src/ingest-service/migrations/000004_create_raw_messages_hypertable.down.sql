SELECT remove_retention_policy('raw_messages', if_exists => true);

DROP TABLE IF EXISTS raw_messages;
