create role asaf login superuser password 'zHsR3PxzVTmr';
create database imageresizer owner asaf;

CREATE TABLE requests (
    uuid UUID PRIMARY KEY NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR (32) NOT NULL
);
CREATE INDEX requests_uuid_index ON requests(uuid);

CREATE TABLE images (
    uuid UUID PRIMARY KEY NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resized_image BYTEA
);
CREATE INDEX images_uuid_index ON images(uuid);