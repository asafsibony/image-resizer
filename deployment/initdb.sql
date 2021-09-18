create role asaf login superuser password 'zHsR3PxzVTmr';
create database imageresizer owner asaf;

CREATE TABLE requests (
    id BIGSERIAL PRIMARY KEY,
    image_uuid UUID NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR (32) NOT NULL
);
CREATE INDEX requests_image_uuid_index ON requests(image_uuid);

CREATE TABLE images (
    uuid UUID PRIMARY KEY NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resized_image BYTEA
);
