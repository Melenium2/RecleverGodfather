create table if not exists  logs (
    message_type String,
    times UInt64,
    log String
) engine = MergeTree() partition by message_type order by (message_type, times)