CREATE TABLE endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(64) NOT NULL,            -- ID организации/пользователя в системе
    url TEXT NOT NULL,                          -- URL эндпоинта клиента (куда слать POST)
    description TEXT,                           -- Человекочитаемое описание
    secret_key VARCHAR(128) NOT NULL,           -- Секрет для HMAC-SHA256 подписи (хешируется/шифруется)
    subscribed_events TEXT[] NOT NULL,          -- Список событий, e.g. ARRAY['order.created', 'payment.succeeded'] или ['*']
    is_active BOOLEAN NOT NULL DEFAULT TRUE,    -- Включен/отключен
    rate_limit_rps INT NOT NULL DEFAULT 50,     -- Лимит запросов в секунду для защиты клиента
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_endpoints_tenant ON endpoints(tenant_id);
CREATE INDEX idx_endpoints_active ON endpoints(is_active) WHERE is_active = TRUE;

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(128) NOT NULL UNIQUE,      -- Идемпотентный ключ события
    event_type VARCHAR(128) NOT NULL,          -- Название события (order.paid, user.registered)
    payload JSONB NOT NULL,                     -- Тело события
    trace_context JSONB,                        -- W3C Trace Context (traceparent, tracestate)
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING', -- PENDING, PUBLISHED, FAILED
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);

-- Частичный индекс для быстрого чтения фоновым релеем (Outbox Worker)
CREATE INDEX idx_outbox_pending ON outbox_events(created_at) 
WHERE status = 'PENDING';

CREATE TABLE delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(128) NOT NULL,
    endpoint_id UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    attempt_number INT NOT NULL DEFAULT 1,      -- Номер попытки (1, 2, 3...)
    status VARCHAR(32) NOT NULL,                -- SUCCESS, FAILED, RETRYING, DEAD_LETTER
    http_status_code INT,                       -- 200, 404, 500, 504 (NULL при таймауте)
    execution_time_ms INT NOT NULL,             -- Задержка HTTP-ответа клиента
    request_headers JSONB,                      -- Отправленные заголовки (включая HMAC сигнатуру)
    response_body TEXT,                         -- Первые 1-2 КБ ответа клиента для дебага
    error_message TEXT,                         -- Текст сетевой ошибки (i/o timeout, connection refused)
    trace_id VARCHAR(64),                       -- Trace ID из OpenTelemetry
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_delivery_endpoint_created ON delivery_attempts(endpoint_id, created_at DESC);
CREATE INDEX idx_delivery_event ON delivery_attempts(event_id);