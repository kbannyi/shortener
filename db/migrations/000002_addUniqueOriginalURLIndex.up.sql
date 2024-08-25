ALTER TABLE public.url
    ADD CONSTRAINT original_url_uniq
        UNIQUE (original_url);