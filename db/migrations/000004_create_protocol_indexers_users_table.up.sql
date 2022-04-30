CREATE TABLE public.protocol_indexers_users (
    user_id                         text NOT NULL,
    protocol_indexer_id        bigserial NOT NULL
);

ALTER TABLE
    ONLY public.protocol_indexers_users
ADD
    CONSTRAINT protocol_indexers_users_pkey
    PRIMARY KEY (user_id, protocol_indexer_id);

ALTER TABLE
    ONLY public.protocol_indexers_users
ADD
    CONSTRAINT protocol_indexers_users_user_id_fkey
    FOREIGN KEY (user_id)
    REFERENCES public.users(address) ON DELETE CASCADE;

ALTER TABLE
    ONLY public.protocol_indexers_users
ADD
    CONSTRAINT protocol_indexers_users_protocol_indexer_id_fkey
    FOREIGN KEY (protocol_indexer_id)
    REFERENCES public.protocol_indexers(id) ON DELETE CASCADE;