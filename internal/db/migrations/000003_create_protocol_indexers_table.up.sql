CREATE TABLE IF NOT EXISTS public.protocol_indexers (
    id              bigserial NOT NULL,
    protocol_id     bigserial NOT NULL,
    name            text NOT NULL,
    spec            jsonb NOT NULL
);

ALTER TABLE
    ONLY public.protocol_indexers
ADD
    CONSTRAINT protocol_indexers_pkey
    PRIMARY KEY (id);

ALTER TABLE
    ONLY public.protocol_indexers
ADD
    CONSTRAINT protocol_indexers_protocol_id_fkey
    FOREIGN KEY (protocol_id)
    REFERENCES public.protocols(id) ON DELETE CASCADE;
