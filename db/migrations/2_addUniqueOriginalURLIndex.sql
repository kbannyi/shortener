ALTER TABLE public.url DROP CONSTRAINT IF EXISTS original_url_uniq;

ALTER TABLE public.url
    ADD CONSTRAINT original_url_uniq
        UNIQUE (original_url);