create role asaf login superuser password 'zHsR3PxzVTmr';
create database resizerdb owner asaf;

CREATE TABLE "image" (
    uuid UUID PRIMARY KEY NOT NULL,
    inserted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resized_image BYTEA,
    status VARCHAR (32) NOT NULL
);

CREATE INDEX  image_id_index ON "image"(uuid);