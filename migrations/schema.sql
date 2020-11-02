--
-- PostgreSQL database dump
--

-- Dumped from database version 10.10 (Debian 10.10-1.pgdg90+1)
-- Dumped by pg_dump version 12.3 (Ubuntu 12.3-1.pgdg19.10+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: climbing_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.climbing_type AS ENUM (
    'TREE',
    'ROCK',
    'BOULDER',
    'ICE',
    'HIGHROPE',
    'OTHER'
);


ALTER TYPE public.climbing_type OWNER TO postgres;

SET default_tablespace = '';

--
-- Name: climbingactivities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.climbingactivities (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    date timestamp without time zone NOT NULL,
    lat numeric NOT NULL,
    lng numeric NOT NULL,
    location character varying(255) NOT NULL,
    type public.climbing_type NOT NULL,
    other_type character varying(255) NOT NULL,
    role character varying(255) NOT NULL,
    comment text NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.climbingactivities OWNER TO postgres;

--
-- Name: participants_climbingactivities; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.participants_climbingactivities (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    climbingactivity_id uuid NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.participants_climbingactivities OWNER TO postgres;

--
-- Name: schema_migration; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migration (
    version character varying(14) NOT NULL
);


ALTER TABLE public.schema_migration OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255),
    provider character varying(255) NOT NULL,
    provider_id character varying(255) NOT NULL,
    avatar_url character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: climbingactivities climbingactivities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.climbingactivities
    ADD CONSTRAINT climbingactivities_pkey PRIMARY KEY (id);


--
-- Name: participants_climbingactivities participants_climbingactivities_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.participants_climbingactivities
    ADD CONSTRAINT participants_climbingactivities_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: schema_migration_version_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX schema_migration_version_idx ON public.schema_migration USING btree (version);


--
-- PostgreSQL database dump complete
--

