--
-- PostgreSQL database cluster dump
--

\restrict B6mMi7e1pQQ0lzevQSRWIcWab6i8Uepn3DWwZsBaxi4WHxpakqfJb6aAfQZWccw

SET default_transaction_read_only = off;

SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;

--
-- Roles
--

CREATE ROLE postgres;
ALTER ROLE postgres WITH SUPERUSER INHERIT CREATEROLE CREATEDB LOGIN REPLICATION BYPASSRLS PASSWORD 'SCRAM-SHA-256$4096:wo7YJNiL85TBEyIcCJw8uA==$36GlBEeWENsWrjwKAmJoNDJK233l+vK799pYuUvRjOk=:lE37httR0ZN0LK9O8ZUwS9E/ce0/u+s3gKTQ1krlU5U=';
CREATE ROLE repluser;
ALTER ROLE repluser WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN REPLICATION NOBYPASSRLS PASSWORD 'SCRAM-SHA-256$4096:VJXcISXzk2GO2YOLgzKJew==$GZVEDdsiLdTgOO9otWqXyMc2S/1ID2/31ymTP4SaCFM=:168yy1St68ucj3WlX+jfCIbAhrpeoQRVaPWoBsfJWqc=';

--
-- User Configurations
--








\unrestrict B6mMi7e1pQQ0lzevQSRWIcWab6i8Uepn3DWwZsBaxi4WHxpakqfJb6aAfQZWccw

--
-- Databases
--

--
-- Database "template1" dump
--

\connect template1

--
-- PostgreSQL database dump
--

\restrict jleeEMC95blR1ffeu5cwyZ5I1BXkP8cM8OpFdYjiZkzdMIch8duvxbit4LKMVgZ

-- Dumped from database version 16.11 (Debian 16.11-1.pgdg13+1)
-- Dumped by pg_dump version 16.11 (Debian 16.11-1.pgdg13+1)

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
-- PostgreSQL database dump complete
--

\unrestrict jleeEMC95blR1ffeu5cwyZ5I1BXkP8cM8OpFdYjiZkzdMIch8duvxbit4LKMVgZ

--
-- Database "postgres" dump
--

\connect postgres

--
-- PostgreSQL database dump
--

\restrict zLi8AsOVNW1TIcBxEuAf2rxw78iqwcye0ZqzAxOz1Cqr5DHLNuHrWs8MijIYYfi

-- Dumped from database version 16.11 (Debian 16.11-1.pgdg13+1)
-- Dumped by pg_dump version 16.11 (Debian 16.11-1.pgdg13+1)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: test_table; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.test_table (
    id integer NOT NULL,
    name text
);


ALTER TABLE public.test_table OWNER TO postgres;

--
-- Name: test_table_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.test_table_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.test_table_id_seq OWNER TO postgres;

--
-- Name: test_table_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.test_table_id_seq OWNED BY public.test_table.id;


--
-- Name: test_table id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_table ALTER COLUMN id SET DEFAULT nextval('public.test_table_id_seq'::regclass);


--
-- Data for Name: test_table; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.test_table (id, name) FROM stdin;
1	Hello from master
\.


--
-- Name: test_table_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.test_table_id_seq', 1, true);


--
-- Name: test_table test_table_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test_table
    ADD CONSTRAINT test_table_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

\unrestrict zLi8AsOVNW1TIcBxEuAf2rxw78iqwcye0ZqzAxOz1Cqr5DHLNuHrWs8MijIYYfi

--
-- PostgreSQL database cluster dump complete
--

