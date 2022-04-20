CREATE TABLE IF NOT EXISTS public.users (address text NOT NULL);

ALTER TABLE
    ONLY public.users
ADD
    CONSTRAINT users_pkey PRIMARY KEY (address);