CREATE TABLE IF NOT EXISTS public.protocols (
    id      bigserial NOT NULL,
    name    text NOT NULL
);

ALTER TABLE
    ONLY public.protocols
ADD
    CONSTRAINT protocols_pkey
    PRIMARY KEY (id);