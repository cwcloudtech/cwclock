-- The AI assistant (ai-instruct-132): the connections a user holds to a language
-- model, who else may spend them, the conversations they have through them, and the
-- catalogue of providers the deployment knows how to talk to. Ported from the
-- uprodit / strong-fish AI feature and kept to the cwclock convention: a uuid id,
-- typed FK + timestamp columns, and everything else in a JSONB `data`.
--
-- FORMAT is one of three WIRE shapes, not a vendor: OpenAI's chat-completions is
-- also what Mistral, DeepSeek and OpenRouter speak, so pointing a connection at a
-- gateway is a change of URL alone. Only Claude and Gemini need a format of their own.

-- ---------------------------------------------------------------------------
-- ai_providers — the catalogue of providers this deployment can talk to.
-- ---------------------------------------------------------------------------
-- A table rather than a switch in the code: a provider is an address, a wire
-- protocol (openai | anthropic | gemini), a default model and the models it serves.
-- All of it moves, and moving it should be an UPDATE, not a release.
CREATE TABLE ai_providers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL UNIQUE,
    format        TEXT NOT NULL,
    base_url      TEXT NOT NULL,
    default_model TEXT NOT NULL,
    -- data: { label, models: [...], keyHint } - what to call it on screen, the
    -- models to suggest (best first), and what its key looks like.
    data          JSONB NOT NULL DEFAULT '{}'::jsonb,
    position      INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The providers cwclock speaks to, with the addresses they default to. base_url /
-- default_model / models / keyHint are starting points offered to the user, never
-- imposed - every field stays editable on a connection.
INSERT INTO ai_providers (name, format, base_url, default_model, position, data) VALUES
    ('openai', 'openai', 'https://api.openai.com/v1', 'gpt-4o-mini', 1, jsonb_build_object(
        'label', 'OpenAI',
        'keyHint', 'sk-...',
        'models', jsonb_build_array('gpt-4.1', 'gpt-4.1-mini', 'gpt-4o', 'gpt-4o-mini', 'o4-mini', 'o3')
    )),
    ('mistral', 'openai', 'https://api.mistral.ai/v1', 'mistral-small-latest', 2, jsonb_build_object(
        'label', 'Mistral',
        'models', jsonb_build_array('mistral-large-latest', 'mistral-medium-latest',
            'mistral-small-latest', 'magistral-medium-latest', 'open-mistral-nemo', 'codestral-latest')
    )),
    ('anthropic', 'anthropic', 'https://api.anthropic.com/v1', 'claude-sonnet-5-5', 3, jsonb_build_object(
        'label', 'Anthropic (Claude)',
        'keyHint', 'sk-ant-...',
        'models', jsonb_build_array('claude-opus-5-5', 'claude-sonnet-5-5', 'claude-haiku-5-5',
            'claude-sonnet-4-5', 'claude-haiku-4-5')
    )),
    ('gemini', 'gemini', 'https://generativelanguage.googleapis.com/v1beta', 'gemini-2.5-flash', 4,
        jsonb_build_object(
        'label', 'Google Gemini',
        'keyHint', 'AIza...',
        'models', jsonb_build_array('gemini-2.5-pro', 'gemini-2.5-flash', 'gemini-2.0-flash')
    )),
    -- One key for every vendor at once, which is why its model names carry the vendor in front.
    ('openrouter', 'openai', 'https://openrouter.ai/api/v1', 'meta-llama/llama-3.3-70b-instruct', 5,
        jsonb_build_object(
        'label', 'OpenRouter',
        'keyHint', 'sk-or-...',
        'models', jsonb_build_array('meta-llama/llama-3.3-70b-instruct', 'anthropic/claude-sonnet-4.5',
            'openai/gpt-4o-mini', 'google/gemini-2.5-flash', 'deepseek/deepseek-chat',
            'qwen/qwen-2.5-72b-instruct')
    )),
    ('deepseek', 'openai', 'https://api.deepseek.com/v1', 'deepseek-chat', 6, jsonb_build_object(
        'label', 'DeepSeek',
        'models', jsonb_build_array('deepseek-chat', 'deepseek-reasoner')
    ));

-- ---------------------------------------------------------------------------
-- ai_connections — one user's access to a model, and the credential behind it.
-- ---------------------------------------------------------------------------
-- The same shape as a shared resource: an account holds the credential and decides
-- who else may spend it. data holds the connection itself - { name, provider,
-- baseUrl, model, maxTokens, token }. The token is a credential sitting in a
-- column; the API never sends it back to a client, only whether one is stored.
-- position orders a user's connections; the lowest is the default a question goes
-- through when nobody picks.
CREATE TABLE ai_connections (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data       JSONB NOT NULL DEFAULT '{}'::jsonb,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_connections_owner ON ai_connections (owner_id, position);

-- ---------------------------------------------------------------------------
-- ai_connection_shares — who may spend somebody else's connection.
-- ---------------------------------------------------------------------------
--   owner - it is theirs: edit it, share it, delete it (implicit, not stored here)
--   user  - may ask the model through it, and nothing else
-- One role per person per connection - granting twice is a change of role.
CREATE TABLE ai_connection_shares (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES ai_connections(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, user_id)
);

CREATE INDEX idx_ai_connection_shares_user ON ai_connection_shares (user_id);

-- ---------------------------------------------------------------------------
-- ai_conversations — a thread of questions and answers.
-- ---------------------------------------------------------------------------
-- Kept because an answer is worth coming back to, and because somebody lending a
-- key out is entitled to see what it was spent on. data: { title } - taken from
-- the first question and renameable.
CREATE TABLE ai_conversations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    data       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_conversations_user ON ai_conversations (user_id, updated_at DESC);

-- ---------------------------------------------------------------------------
-- ai_prompts — one exchange: a question and the answer it got.
-- ---------------------------------------------------------------------------
-- A row is one *exchange*, question and answer together. data: { provider, model,
-- connection, usage } - what produced the answer and what it cost. The system
-- prompt is deliberately not stored: it lives in a file and is meant to improve
-- underneath old conversations.
CREATE TABLE ai_prompts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    question        TEXT NOT NULL,
    answer          TEXT NOT NULL,
    data            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_prompts_conversation ON ai_prompts (conversation_id, created_at);
