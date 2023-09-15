CREATE TABLE IF NOT EXISTS openline_carrier (
    id SERIAL PRIMARY KEY NOT NULL,
    carrier_name VARCHAR(50) NOT NULL,
    username VARCHAR(50) NOT NULL,
    realm VARCHAR(50) NOT NULL,
    ha1 VARCHAR(50) NOT NULL,
    domain VARCHAR(250) NOT NULL
);

CREATE UNIQUE INDEX carrier_name_idx ON openline_carrier (carrier_name);

CREATE TABLE IF NOT EXISTS openline_profile
(
    id SERIAL PRIMARY KEY NOT NULL,
    profile_name VARCHAR(50) NOT NULL,
    call_webhook VARCHAR(250) NOT NULL,
    recording_webhook VARCHAR(250) NOT NULL,
    api_key VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS openline_voicemail
(
    id SERIAL PRIMARY KEY NOT NULL,
    object_id VARCHAR(50) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    timeout INT NOT NULL DEFAULT 15
);

CREATE TABLE IF NOT EXISTS openline_forwarding
(
    id SERIAL PRIMARY KEY NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    e164 VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS openline_number_mapping (
    id SERIAL PRIMARY KEY NOT NULL,
    e164 VARCHAR(50) NOT NULL,
    sipuri VARCHAR(250) NOT NULL,
    carrier_name VARCHAR(50) NOT NULL,
    alias VARCHAR(50) NOT NULL,
    phoneuri VARCHAR(250) NOT NULL DEFAULT '',
    profile_id INT,
    voicemail_id INT,
    forwarding_id INT,
    CONSTRAINT fk_profile_id
        FOREIGN KEY(profile_id)
            REFERENCES openline_profile(id),
    CONSTRAINT fk_voicemail_id
        FOREIGN KEY(voicemail_id)
            REFERENCES openline_voicemail(id),
    CONSTRAINT fk_forwarding_id
        FOREIGN KEY(forwarding_id)
            REFERENCES openline_forwarding(id)
);

CREATE UNIQUE INDEX number_e164_idx ON openline_number_mapping (e164);
CREATE UNIQUE INDEX number_sipuri_idx ON openline_number_mapping (sipuri);


