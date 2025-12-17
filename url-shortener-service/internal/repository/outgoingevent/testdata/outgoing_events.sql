TRUNCATE TABLE outgoing_events RESTART IDENTITY;

INSERT INTO outgoing_events (id, topic, correlation_id, trace_id, span_id, payload, status, created_at, updated_at)
VALUES
    (3, 'evt.pending.3', 'a1b2c3', 'xxx', 'yyy', '{"event_id":3,"data":{"short_code":"3","original_url":"foo.bar/3"}}', 'PENDING', NOW(), NOW()),
    (1,  'evt.pending.1', '111222','bbb','ccc',  '{"event_id":1,"data":{"short_code":"1","original_url":"foo.bar/1"}}', 'PENDING', NOW(), NOW()),
    (2,  'evt.pending.2', 'qwerty', '999','888', '{"event_id":2,"data":{"short_code":"2","original_url":"foo.bar/2"}}', 'PENDING', NOW(), NOW()),
    (10, 'evt.sent.10', 'test_123', 'aaa','zzz','{"event_id":10}',                  'SENT',    NOW(), NOW());
