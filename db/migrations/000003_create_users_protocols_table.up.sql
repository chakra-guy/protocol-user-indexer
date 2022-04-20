CREATE TABLE public.users_protocols (
    user_id text NOT NULL,
    protocol_id bigserial NOT NULL,
    interaction_count integer NOT NULL DEFAULT 0
);

ALTER TABLE
    ONLY public.users_protocols
ADD
    CONSTRAINT users_protocols_pkey PRIMARY KEY (user_id, protocol_id);

ALTER TABLE
    ONLY public.users_protocols
ADD
    CONSTRAINT users_protocols_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(address) ON DELETE CASCADE;

ALTER TABLE
    ONLY public.users_protocols
ADD
    CONSTRAINT users_protocols_protocol_id_fkey FOREIGN KEY (protocol_id) REFERENCES public.protocols(id) ON DELETE CASCADE;