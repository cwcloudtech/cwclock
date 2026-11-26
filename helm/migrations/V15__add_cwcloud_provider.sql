-- CWCloud (CWAI) as an AI provider. CWCloud exposes an
-- OpenAI-unlike prompt API: POST {base}/v1/ai/prompt with an X-Auth-Token
-- header, the model carried in an `adapter` field, turns whose text is in
-- `message`, and a 201 reply whose `response` is a string or an array of them.
-- It is a new wire `format` ("cwcloud") handled in internal/ai; the providers
-- table has no CHECK on format, so this is a plain INSERT.
--
-- base_url defaults to the full endpoint because that is the line CWCloud's docs
-- give people to paste; the client also accepts the host alone or the /v1 root
-- and appends what is missing. The models are CWCloud's adapter names.
--
-- Unlike every other provider, CWAI cannot do tools, so the assistant asks it
-- with the tool catalogue withheld (see models.AIFormatSupportsTools).
INSERT INTO ai_providers (name, format, base_url, default_model, position, data) VALUES
    ('cwcloud', 'cwcloud', 'https://api.cwcloud.tech/v1/ai/prompt', 'claude-sonnet', 7, jsonb_build_object(
        'label', 'CWCloud (CWAI)',
        'keyHint', 'the CWCloud API token',
        'models', jsonb_build_array('mistral-medium', 'claude-sonnet', 'claude-haiku', 'gemini-flash')
    ));
