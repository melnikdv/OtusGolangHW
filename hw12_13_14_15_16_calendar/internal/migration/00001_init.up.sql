CREATE TABLE IF NOT EXISTS events (
                                      id TEXT PRIMARY KEY,
                                      title TEXT NOT NULL,
                                      datetime TIMESTAMP WITH TIME ZONE NOT NULL,
                                      duration BIGINT NOT NULL,
                                      description TEXT,
                                      user_id TEXT NOT NULL,
                                      notify_before BIGINT,
                                      notified BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_events_user_datetime ON events (user_id, datetime);
