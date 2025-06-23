--
-- PostgreSQL database dump
--

-- Dumped from database version 17.5 (Debian 17.5-1.pgdg120+1)
-- Dumped by pg_dump version 17.5 (Debian 17.5-1.pgdg120+1)

-- Started on 2025-06-21 09:37:25 UTC

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3810 (class 0 OID 0)
-- Dependencies: 6
-- Name: SCHEMA "public"; Type: COMMENT; Schema: -; Owner: pg_database_owner
--

COMMENT ON SCHEMA "public" IS 'standard public schema';


--
-- TOC entry 3 (class 3079 OID 25169)
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "public";


--
-- TOC entry 3811 (class 0 OID 0)
-- Dependencies: 3
-- Name: EXTENSION "pgcrypto"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "pgcrypto" IS 'cryptographic functions';


--
-- TOC entry 2 (class 3079 OID 25158)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "public";


--
-- TOC entry 3812 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- TOC entry 1000 (class 1247 OID 25732)
-- Name: conversation_type; Type: TYPE; Schema: public; Owner: veza
--

CREATE TYPE "public"."conversation_type" AS ENUM (
    'direct_message',
    'public_room',
    'private_room',
    'group'
);


ALTER TYPE "public"."conversation_type" OWNER TO "veza";

--
-- TOC entry 997 (class 1247 OID 25720)
-- Name: message_status; Type: TYPE; Schema: public; Owner: veza
--

CREATE TYPE "public"."message_status" AS ENUM (
    'sent',
    'delivered',
    'read',
    'edited',
    'deleted'
);


ALTER TYPE "public"."message_status" OWNER TO "veza";

--
-- TOC entry 994 (class 1247 OID 25711)
-- Name: user_role; Type: TYPE; Schema: public; Owner: veza
--

CREATE TYPE "public"."user_role" AS ENUM (
    'user',
    'moderator',
    'admin',
    'super_admin'
);


ALTER TYPE "public"."user_role" OWNER TO "veza";

--
-- TOC entry 267 (class 1255 OID 25871)
-- Name: cleanup_expired_sessions(); Type: FUNCTION; Schema: public; Owner: veza
--

CREATE FUNCTION "public"."cleanup_expired_sessions"() RETURNS integer
    LANGUAGE "plpgsql"
    AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions 
    WHERE expires_at < NOW() OR (last_activity < NOW() - INTERVAL '30 days');
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$;


ALTER FUNCTION "public"."cleanup_expired_sessions"() OWNER TO "veza";

--
-- TOC entry 265 (class 1255 OID 25870)
-- Name: current_user_id(); Type: FUNCTION; Schema: public; Owner: veza
--

CREATE FUNCTION "public"."current_user_id"() RETURNS bigint
    LANGUAGE "plpgsql" SECURITY DEFINER
    AS $$
BEGIN
    -- Cette fonction doit être implémentée côté application
    -- Pour l'instant, elle retourne NULL
    RETURN NULL;
END;
$$;


ALTER FUNCTION "public"."current_user_id"() OWNER TO "veza";

--
-- TOC entry 266 (class 1255 OID 24860)
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: veza
--

CREATE FUNCTION "public"."update_updated_at_column"() RETURNS "trigger"
    LANGUAGE "plpgsql"
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION "public"."update_updated_at_column"() OWNER TO "veza";

SET default_tablespace = '';

SET default_table_access_method = "heap";

--
-- TOC entry 256 (class 1259 OID 25120)
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."audit_logs" (
    "id" integer NOT NULL,
    "user_id" integer,
    "action" character varying(100) NOT NULL,
    "resource_type" character varying(50),
    "resource_id" character varying(100),
    "details" "jsonb",
    "ip_address" "inet",
    "user_agent" "text",
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE "public"."audit_logs" OWNER TO "veza";

--
-- TOC entry 255 (class 1259 OID 25119)
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."audit_logs_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."audit_logs_id_seq" OWNER TO "veza";

--
-- TOC entry 3813 (class 0 OID 0)
-- Dependencies: 255
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."audit_logs_id_seq" OWNED BY "public"."audit_logs"."id";


--
-- TOC entry 260 (class 1259 OID 25798)
-- Name: conversation_members; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."conversation_members" (
    "id" bigint NOT NULL,
    "conversation_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "role" character varying(20) DEFAULT 'member'::character varying NOT NULL,
    "joined_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "left_at" timestamp with time zone,
    "is_muted" boolean DEFAULT false
);


ALTER TABLE "public"."conversation_members" OWNER TO "veza";

--
-- TOC entry 259 (class 1259 OID 25797)
-- Name: conversation_members_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."conversation_members_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."conversation_members_id_seq" OWNER TO "veza";

--
-- TOC entry 3814 (class 0 OID 0)
-- Dependencies: 259
-- Name: conversation_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."conversation_members_id_seq" OWNED BY "public"."conversation_members"."id";


--
-- TOC entry 258 (class 1259 OID 25757)
-- Name: conversations; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."conversations" (
    "id" bigint NOT NULL,
    "uuid" "uuid" DEFAULT "public"."uuid_generate_v4"() NOT NULL,
    "type" "public"."conversation_type" DEFAULT 'direct_message'::"public"."conversation_type" NOT NULL,
    "name" character varying(100),
    "description" "text",
    "owner_id" bigint NOT NULL,
    "is_public" boolean DEFAULT false NOT NULL,
    "is_archived" boolean DEFAULT false NOT NULL,
    "max_members" integer DEFAULT 100,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."conversations" OWNER TO "veza";

--
-- TOC entry 257 (class 1259 OID 25756)
-- Name: conversations_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."conversations_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."conversations_id_seq" OWNER TO "veza";

--
-- TOC entry 3815 (class 0 OID 0)
-- Dependencies: 257
-- Name: conversations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."conversations_id_seq" OWNED BY "public"."conversations"."id";


--
-- TOC entry 224 (class 1259 OID 16436)
-- Name: files; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."files" (
    "id" integer NOT NULL,
    "product_id" integer NOT NULL,
    "filename" "text" NOT NULL,
    "url" "text" NOT NULL,
    "type" "text" NOT NULL,
    "uploaded_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "public"."files" OWNER TO "veza";

--
-- TOC entry 223 (class 1259 OID 16435)
-- Name: files_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."files_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."files_id_seq" OWNER TO "veza";

--
-- TOC entry 3816 (class 0 OID 0)
-- Dependencies: 223
-- Name: files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."files_id_seq" OWNED BY "public"."files"."id";


--
-- TOC entry 264 (class 1259 OID 25841)
-- Name: message_history; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."message_history" (
    "id" bigint NOT NULL,
    "message_id" bigint NOT NULL,
    "old_content" "text" NOT NULL,
    "edited_by" bigint NOT NULL,
    "edited_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."message_history" OWNER TO "veza";

--
-- TOC entry 263 (class 1259 OID 25840)
-- Name: message_history_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."message_history_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."message_history_id_seq" OWNER TO "veza";

--
-- TOC entry 3817 (class 0 OID 0)
-- Dependencies: 263
-- Name: message_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."message_history_id_seq" OWNED BY "public"."message_history"."id";


--
-- TOC entry 262 (class 1259 OID 25820)
-- Name: message_mentions; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."message_mentions" (
    "id" bigint NOT NULL,
    "message_id" bigint NOT NULL,
    "mentioned_user_id" bigint NOT NULL,
    "is_read" boolean DEFAULT false,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."message_mentions" OWNER TO "veza";

--
-- TOC entry 261 (class 1259 OID 25819)
-- Name: message_mentions_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."message_mentions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."message_mentions_id_seq" OWNER TO "veza";

--
-- TOC entry 3818 (class 0 OID 0)
-- Dependencies: 261
-- Name: message_mentions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."message_mentions_id_seq" OWNED BY "public"."message_mentions"."id";


--
-- TOC entry 246 (class 1259 OID 25008)
-- Name: message_reactions; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."message_reactions" (
    "id" integer NOT NULL,
    "message_id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "reaction_type" character varying(100) NOT NULL,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "emoji" character varying(20) NOT NULL
);


ALTER TABLE "public"."message_reactions" OWNER TO "veza";

--
-- TOC entry 3819 (class 0 OID 0)
-- Dependencies: 246
-- Name: TABLE "message_reactions"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."message_reactions" IS 'Table des réactions aux messages (like, love, etc.)';


--
-- TOC entry 245 (class 1259 OID 25007)
-- Name: message_reactions_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."message_reactions_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."message_reactions_id_seq" OWNER TO "veza";

--
-- TOC entry 3820 (class 0 OID 0)
-- Dependencies: 245
-- Name: message_reactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."message_reactions_id_seq" OWNED BY "public"."message_reactions"."id";


--
-- TOC entry 228 (class 1259 OID 16479)
-- Name: messages; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."messages" (
    "id" integer NOT NULL,
    "author_id" integer,
    "to_user" integer,
    "room" "text",
    "content" "text" NOT NULL,
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "message_type" character varying(20) DEFAULT 'text'::character varying,
    "reply_to_id" integer,
    "is_edited" boolean DEFAULT false,
    "edited_at" timestamp with time zone,
    "metadata" "jsonb",
    "is_pinned" boolean DEFAULT false,
    "thread_count" integer DEFAULT 0,
    "status" character varying(20) DEFAULT 'sent'::character varying,
    "uuid" "uuid" DEFAULT "public"."uuid_generate_v4"(),
    "conversation_id" bigint,
    "parent_message_id" bigint,
    "updated_at" timestamp with time zone DEFAULT "now"(),
    CONSTRAINT "chk_message_content_length" CHECK (("length"("content") <= 4000))
);


ALTER TABLE "public"."messages" OWNER TO "veza";

--
-- TOC entry 227 (class 1259 OID 16478)
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."messages_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."messages_id_seq" OWNER TO "veza";

--
-- TOC entry 3821 (class 0 OID 0)
-- Dependencies: 227
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."messages_id_seq" OWNED BY "public"."messages"."id";


--
-- TOC entry 238 (class 1259 OID 24885)
-- Name: migrations; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."migrations" (
    "id" integer NOT NULL,
    "filename" character varying(255) NOT NULL,
    "applied_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "public"."migrations" OWNER TO "veza";

--
-- TOC entry 237 (class 1259 OID 24884)
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."migrations_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."migrations_id_seq" OWNER TO "veza";

--
-- TOC entry 3822 (class 0 OID 0)
-- Dependencies: 237
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."migrations_id_seq" OWNED BY "public"."migrations"."id";


--
-- TOC entry 252 (class 1259 OID 25078)
-- Name: notifications; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."notifications" (
    "id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "type" character varying(50) NOT NULL,
    "title" character varying(255) NOT NULL,
    "content" "text" NOT NULL,
    "metadata" "jsonb",
    "is_read" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "read_at" timestamp with time zone
);


ALTER TABLE "public"."notifications" OWNER TO "veza";

--
-- TOC entry 3823 (class 0 OID 0)
-- Dependencies: 252
-- Name: TABLE "notifications"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."notifications" IS 'Table des notifications push/in-app';


--
-- TOC entry 251 (class 1259 OID 25077)
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."notifications_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."notifications_id_seq" OWNER TO "veza";

--
-- TOC entry 3824 (class 0 OID 0)
-- Dependencies: 251
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."notifications_id_seq" OWNED BY "public"."notifications"."id";


--
-- TOC entry 232 (class 1259 OID 24729)
-- Name: offers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE "public"."offers" (
    "id" integer NOT NULL,
    "listing_id" integer NOT NULL,
    "from_user_id" integer NOT NULL,
    "proposed_product_id" integer NOT NULL,
    "message" "text",
    "status" "text" DEFAULT 'pending'::"text" NOT NULL,
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "public"."offers" OWNER TO "postgres";

--
-- TOC entry 231 (class 1259 OID 24728)
-- Name: offers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE "public"."offers_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."offers_id_seq" OWNER TO "postgres";

--
-- TOC entry 3826 (class 0 OID 0)
-- Dependencies: 231
-- Name: offers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE "public"."offers_id_seq" OWNED BY "public"."offers"."id";


--
-- TOC entry 236 (class 1259 OID 24841)
-- Name: product_documents; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."product_documents" (
    "id" integer NOT NULL,
    "product_id" integer NOT NULL,
    "name" "text" NOT NULL,
    "description" "text" DEFAULT ''::"text",
    "file_type" "text" NOT NULL,
    "file_path" "text" NOT NULL,
    "file_size" bigint DEFAULT 0,
    "uploaded_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "product_documents_file_type_check" CHECK (("file_type" = ANY (ARRAY['manual'::"text", 'datasheet'::"text", 'warranty'::"text", 'image'::"text", 'other'::"text"])))
);


ALTER TABLE "public"."product_documents" OWNER TO "veza";

--
-- TOC entry 235 (class 1259 OID 24840)
-- Name: product_documents_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."product_documents_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."product_documents_id_seq" OWNER TO "veza";

--
-- TOC entry 3828 (class 0 OID 0)
-- Dependencies: 235
-- Name: product_documents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."product_documents_id_seq" OWNED BY "public"."product_documents"."id";


--
-- TOC entry 234 (class 1259 OID 24771)
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE "public"."products" (
    "id" integer NOT NULL,
    "name" "text" NOT NULL,
    "category_id" integer,
    "brand" "text" DEFAULT ''::"text",
    "model" "text" DEFAULT ''::"text",
    "description" "text" DEFAULT ''::"text",
    "price" numeric(10,2),
    "warranty_months" integer DEFAULT 0,
    "warranty_conditions" "text" DEFAULT ''::"text",
    "manufacturer_website" "text" DEFAULT ''::"text",
    "specifications" "text" DEFAULT ''::"text",
    "status" "text" DEFAULT 'active'::"text",
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT "products_status_check" CHECK (("status" = ANY (ARRAY['active'::"text", 'discontinued'::"text", 'draft'::"text"])))
);


ALTER TABLE "public"."products" OWNER TO "postgres";

--
-- TOC entry 233 (class 1259 OID 24770)
-- Name: products_id_seq1; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE "public"."products_id_seq1"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."products_id_seq1" OWNER TO "postgres";

--
-- TOC entry 3830 (class 0 OID 0)
-- Dependencies: 233
-- Name: products_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE "public"."products_id_seq1" OWNED BY "public"."products"."id";


--
-- TOC entry 242 (class 1259 OID 24956)
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE "public"."refresh_tokens" (
    "id" integer NOT NULL,
    "user_id" integer,
    "token" "text" NOT NULL,
    "expires_at" timestamp without time zone NOT NULL,
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE "public"."refresh_tokens" OWNER TO "postgres";

--
-- TOC entry 241 (class 1259 OID 24955)
-- Name: refresh_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE "public"."refresh_tokens_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."refresh_tokens_id_seq" OWNER TO "postgres";

--
-- TOC entry 3833 (class 0 OID 0)
-- Dependencies: 241
-- Name: refresh_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE "public"."refresh_tokens_id_seq" OWNED BY "public"."refresh_tokens"."id";


--
-- TOC entry 250 (class 1259 OID 25054)
-- Name: room_members; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."room_members" (
    "id" integer NOT NULL,
    "room_id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "role" character varying(20) DEFAULT 'member'::character varying NOT NULL,
    "joined_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "last_read_at" timestamp with time zone
);


ALTER TABLE "public"."room_members" OWNER TO "veza";

--
-- TOC entry 3835 (class 0 OID 0)
-- Dependencies: 250
-- Name: TABLE "room_members"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."room_members" IS 'Table des membres de salon avec leurs rôles';


--
-- TOC entry 249 (class 1259 OID 25053)
-- Name: room_members_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."room_members_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."room_members_id_seq" OWNER TO "veza";

--
-- TOC entry 3836 (class 0 OID 0)
-- Dependencies: 249
-- Name: room_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."room_members_id_seq" OWNED BY "public"."room_members"."id";


--
-- TOC entry 226 (class 1259 OID 16466)
-- Name: rooms; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."rooms" (
    "id" integer NOT NULL,
    "name" "text" NOT NULL,
    "is_private" boolean DEFAULT false,
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "creator_id" integer,
    "max_members" integer DEFAULT 1000,
    "description" "text",
    CONSTRAINT "chk_room_name_length" CHECK ((("length"("name") <= 50) AND ("length"("name") >= 1)))
);


ALTER TABLE "public"."rooms" OWNER TO "veza";

--
-- TOC entry 3837 (class 0 OID 0)
-- Dependencies: 226
-- Name: TABLE "rooms"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."rooms" IS 'Table des salons de chat avec métadonnées';


--
-- TOC entry 225 (class 1259 OID 16465)
-- Name: rooms_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."rooms_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."rooms_id_seq" OWNER TO "veza";

--
-- TOC entry 3838 (class 0 OID 0)
-- Dependencies: 225
-- Name: rooms_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."rooms_id_seq" OWNED BY "public"."rooms"."id";


--
-- TOC entry 244 (class 1259 OID 24985)
-- Name: sanctions; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."sanctions" (
    "id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "moderator_id" integer NOT NULL,
    "sanction_type" character varying(50) NOT NULL,
    "reason" character varying(100) NOT NULL,
    "message" "text",
    "expires_at" timestamp with time zone,
    "is_active" boolean DEFAULT true NOT NULL,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE "public"."sanctions" OWNER TO "veza";

--
-- TOC entry 3839 (class 0 OID 0)
-- Dependencies: 244
-- Name: TABLE "sanctions"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."sanctions" IS 'Table des sanctions de modération (warnings, mutes, bans)';


--
-- TOC entry 243 (class 1259 OID 24984)
-- Name: sanctions_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."sanctions_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."sanctions_id_seq" OWNER TO "veza";

--
-- TOC entry 3840 (class 0 OID 0)
-- Dependencies: 243
-- Name: sanctions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."sanctions_id_seq" OWNED BY "public"."sanctions"."id";


--
-- TOC entry 230 (class 1259 OID 16550)
-- Name: tags; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE "public"."tags" (
    "id" integer NOT NULL,
    "name" "text" NOT NULL
);


ALTER TABLE "public"."tags" OWNER TO "postgres";

--
-- TOC entry 229 (class 1259 OID 16549)
-- Name: tags_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE "public"."tags_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."tags_id_seq" OWNER TO "postgres";

--
-- TOC entry 3842 (class 0 OID 0)
-- Dependencies: 229
-- Name: tags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE "public"."tags_id_seq" OWNED BY "public"."tags"."id";


--
-- TOC entry 248 (class 1259 OID 25029)
-- Name: user_blocks; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."user_blocks" (
    "id" integer NOT NULL,
    "blocker_id" integer NOT NULL,
    "blocked_id" integer NOT NULL,
    "reason" character varying(255),
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT "user_blocks_check" CHECK (("blocker_id" <> "blocked_id"))
);


ALTER TABLE "public"."user_blocks" OWNER TO "veza";

--
-- TOC entry 3843 (class 0 OID 0)
-- Dependencies: 248
-- Name: TABLE "user_blocks"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."user_blocks" IS 'Table des blocages entre utilisateurs';


--
-- TOC entry 247 (class 1259 OID 25028)
-- Name: user_blocks_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."user_blocks_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."user_blocks_id_seq" OWNER TO "veza";

--
-- TOC entry 3844 (class 0 OID 0)
-- Dependencies: 247
-- Name: user_blocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."user_blocks_id_seq" OWNED BY "public"."user_blocks"."id";


--
-- TOC entry 254 (class 1259 OID 25097)
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."user_sessions" (
    "id" integer NOT NULL,
    "user_id" integer NOT NULL,
    "session_token" character varying(255) NOT NULL,
    "device_info" character varying(255),
    "ip_address" "inet",
    "last_activity" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "expires_at" timestamp with time zone NOT NULL,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "is_active" boolean DEFAULT true NOT NULL
);


ALTER TABLE "public"."user_sessions" OWNER TO "veza";

--
-- TOC entry 3845 (class 0 OID 0)
-- Dependencies: 254
-- Name: TABLE "user_sessions"; Type: COMMENT; Schema: public; Owner: veza
--

COMMENT ON TABLE "public"."user_sessions" IS 'Table des sessions utilisateur actives';


--
-- TOC entry 253 (class 1259 OID 25096)
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."user_sessions_id_seq"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."user_sessions_id_seq" OWNER TO "veza";

--
-- TOC entry 3846 (class 0 OID 0)
-- Dependencies: 253
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."user_sessions_id_seq" OWNED BY "public"."user_sessions"."id";


--
-- TOC entry 240 (class 1259 OID 24924)
-- Name: users; Type: TABLE; Schema: public; Owner: veza
--

CREATE TABLE "public"."users" (
    "id" integer NOT NULL,
    "username" "text" NOT NULL,
    "email" "text" NOT NULL,
    "password_hash" "text" NOT NULL,
    "first_name" "text" DEFAULT ''::"text",
    "last_name" "text" DEFAULT ''::"text",
    "avatar" "text" DEFAULT ''::"text",
    "bio" "text" DEFAULT ''::"text",
    "role" "text" DEFAULT 'user'::"text",
    "is_active" boolean DEFAULT true,
    "is_verified" boolean DEFAULT false,
    "last_login_at" timestamp without time zone,
    "created_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    "status" character varying(20) DEFAULT 'offline'::character varying,
    "last_seen" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    "reputation_score" integer DEFAULT 100,
    "is_banned" boolean DEFAULT false,
    "is_muted" boolean DEFAULT false,
    "uuid" "uuid" DEFAULT "public"."uuid_generate_v4"(),
    "last_activity" timestamp with time zone DEFAULT "now"(),
    CONSTRAINT "users_role_check" CHECK (("role" = ANY (ARRAY['user'::"text", 'admin'::"text", 'super_admin'::"text", 'moderator'::"text"])))
);


ALTER TABLE "public"."users" OWNER TO "veza";

--
-- TOC entry 239 (class 1259 OID 24923)
-- Name: users_id_seq1; Type: SEQUENCE; Schema: public; Owner: veza
--

CREATE SEQUENCE "public"."users_id_seq1"
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."users_id_seq1" OWNER TO "veza";

--
-- TOC entry 3847 (class 0 OID 0)
-- Dependencies: 239
-- Name: users_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: veza
--

ALTER SEQUENCE "public"."users_id_seq1" OWNED BY "public"."users"."id";


--
-- TOC entry 3448 (class 2604 OID 25123)
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."audit_logs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."audit_logs_id_seq"'::"regclass");


--
-- TOC entry 3458 (class 2604 OID 25801)
-- Name: conversation_members id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversation_members" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."conversation_members_id_seq"'::"regclass");


--
-- TOC entry 3450 (class 2604 OID 25760)
-- Name: conversations id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversations" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."conversations_id_seq"'::"regclass");


--
-- TOC entry 3376 (class 2604 OID 16439)
-- Name: files id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."files" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."files_id_seq"'::"regclass");


--
-- TOC entry 3465 (class 2604 OID 25844)
-- Name: message_history id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_history" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."message_history_id_seq"'::"regclass");


--
-- TOC entry 3462 (class 2604 OID 25823)
-- Name: message_mentions id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_mentions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."message_mentions_id_seq"'::"regclass");


--
-- TOC entry 3434 (class 2604 OID 25011)
-- Name: message_reactions id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_reactions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."message_reactions_id_seq"'::"regclass");


--
-- TOC entry 3382 (class 2604 OID 16482)
-- Name: messages id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."messages_id_seq"'::"regclass");


--
-- TOC entry 3410 (class 2604 OID 24888)
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."migrations" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."migrations_id_seq"'::"regclass");


--
-- TOC entry 3441 (class 2604 OID 25081)
-- Name: notifications id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."notifications" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."notifications_id_seq"'::"regclass");


--
-- TOC entry 3392 (class 2604 OID 24732)
-- Name: offers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."offers" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."offers_id_seq"'::"regclass");


--
-- TOC entry 3406 (class 2604 OID 24844)
-- Name: product_documents id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."product_documents" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."product_documents_id_seq"'::"regclass");


--
-- TOC entry 3395 (class 2604 OID 24774)
-- Name: products id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."products" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."products_id_seq1"'::"regclass");


--
-- TOC entry 3429 (class 2604 OID 24959)
-- Name: refresh_tokens id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."refresh_tokens" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."refresh_tokens_id_seq"'::"regclass");


--
-- TOC entry 3438 (class 2604 OID 25057)
-- Name: room_members id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."room_members" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."room_members_id_seq"'::"regclass");


--
-- TOC entry 3378 (class 2604 OID 16469)
-- Name: rooms id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."rooms" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."rooms_id_seq"'::"regclass");


--
-- TOC entry 3431 (class 2604 OID 24988)
-- Name: sanctions id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."sanctions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."sanctions_id_seq"'::"regclass");


--
-- TOC entry 3391 (class 2604 OID 16553)
-- Name: tags id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."tags" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."tags_id_seq"'::"regclass");


--
-- TOC entry 3436 (class 2604 OID 25032)
-- Name: user_blocks id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_blocks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."user_blocks_id_seq"'::"regclass");


--
-- TOC entry 3444 (class 2604 OID 25100)
-- Name: user_sessions id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."user_sessions_id_seq"'::"regclass");


--
-- TOC entry 3412 (class 2604 OID 24927)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."users" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."users_id_seq1"'::"regclass");


--
-- TOC entry 3796 (class 0 OID 25120)
-- Dependencies: 256
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3800 (class 0 OID 25798)
-- Dependencies: 260
-- Data for Name: conversation_members; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3798 (class 0 OID 25757)
-- Dependencies: 258
-- Data for Name: conversations; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."conversations" ("id", "uuid", "type", "name", "description", "owner_id", "is_public", "is_archived", "max_members", "created_at", "updated_at") VALUES (1, '786ecc3d-0beb-4c5c-98ad-6970a7e8459a', 'public_room', 'afterworks', NULL, 1, true, false, 100, '2025-06-18 23:48:03.297183+00', '2025-06-18 23:48:03.297183+00');
INSERT INTO "public"."conversations" ("id", "uuid", "type", "name", "description", "owner_id", "is_public", "is_archived", "max_members", "created_at", "updated_at") VALUES (2, 'd6798cce-cca2-40ec-9884-6ad0b52f3cb4', 'public_room', 'general', NULL, 1, true, false, 100, '2025-06-18 23:48:03.297183+00', '2025-06-18 23:48:03.297183+00');
INSERT INTO "public"."conversations" ("id", "uuid", "type", "name", "description", "owner_id", "is_public", "is_archived", "max_members", "created_at", "updated_at") VALUES (3, '933472ec-aca8-4daf-8491-3915fdd70c0b', 'public_room', 'general', NULL, 1, true, false, 100, '2025-06-18 23:56:00.769525+00', '2025-06-18 23:56:00.769525+00');


--
-- TOC entry 3764 (class 0 OID 16436)
-- Dependencies: 224
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."files" ("id", "product_id", "filename", "url", "type", "uploaded_at") VALUES (1, 1, 'test_upload.txt', '/files/1_1747129598_test_upload.txt', 'test', '2025-05-13 09:46:38.538325');


--
-- TOC entry 3804 (class 0 OID 25841)
-- Dependencies: 264
-- Data for Name: message_history; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3802 (class 0 OID 25820)
-- Dependencies: 262
-- Data for Name: message_mentions; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3786 (class 0 OID 25008)
-- Dependencies: 246
-- Data for Name: message_reactions; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3768 (class 0 OID 16479)
-- Dependencies: 228
-- Data for Name: messages; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (61, 6, 5, NULL, 'test', '2025-05-17 16:02:19.936134', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '211dd2bc-aa3a-417b-ab7f-ccc640313a50', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (62, 5, 6, NULL, 'test', '2025-05-17 16:02:30.820598', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'a84f5a3b-5ee8-48eb-a597-d33bbb81ed02', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (63, 6, 5, NULL, 'incroybalke', '2025-05-17 16:02:34.624062', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '0c01ff28-4388-458e-92a2-33e3d037d9a6', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (64, 5, 6, NULL, 'ca fonctionne on dirait', '2025-05-17 16:02:41.124823', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '41fd0d2a-64da-4a42-85c9-6991499994bb', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (65, 6, 5, NULL, 'oui apparememnt', '2025-05-17 16:02:46.639404', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '5bb03da3-499b-4e99-bda7-dfcd1d9e7477', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (66, 5, 6, NULL, 'ok on voit si ca arche', '2025-05-17 16:07:22.176381', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '5e76912d-04a3-4537-a559-4d497ce9834b', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (67, 6, 5, NULL, 'ca a pas l''air', '2025-05-17 16:07:34.065075', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'c6466af6-3ab2-48ea-a54d-b1f9911c4d50', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (68, 5, 6, NULL, 'il y a que l''autre', '2025-05-17 16:07:39.463208', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '67f32415-704d-465e-affd-6209386aee38', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (69, 6, 5, NULL, 'test', '2025-05-17 16:14:16.014101', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e04b7637-a02e-4f7a-9242-a89ff77d7b59', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (70, 5, 6, NULL, 'avion', '2025-05-17 16:14:23.268292', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '5be3c331-6551-4eb6-8597-c7ac68c13e51', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (71, 6, 5, NULL, 'à réaction ?', '2025-05-17 16:14:28.024603', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b391da0e-ff93-419c-be5a-4b3e72a4e863', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (93, 5, 7, NULL, 'test avec biddie', '2025-05-30 15:02:16.086658', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '6382a77f-5800-47f0-bb5f-3634d3d0fcc7', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (94, 5, 6, NULL, 'test avec toto', '2025-05-30 15:02:22.319899', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '6a66e698-7a63-484e-bf81-a23472a33c4f', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (95, 5, 3, NULL, 'test avec testuser', '2025-05-30 15:02:28.963108', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '985355ce-e96e-4fd2-8a10-32187b9e5ffd', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (96, 11, 5, NULL, 'ca va ou quoi', '2025-05-30 15:12:31.868576', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'c87d13da-0b40-400b-a650-d5b73bb89fa3', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (97, 5, 11, NULL, 'oui trkl et toi', '2025-05-30 15:12:43.243662', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '03090109-2c8d-46b7-9266-4363da09ce7a', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (98, 11, 5, NULL, 'bah je vois pas nos ancieen messages donc non', '2025-05-30 15:12:54.490768', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1c695b24-b262-4964-b1a3-4e1e4113e873', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (99, 5, 6, NULL, 'jshdsqdq', '2025-05-30 15:31:37.933694', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '613c6fa5-16e1-418a-832a-005ec5501e1d', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (100, 5, 6, NULL, 'qsdjqskd', '2025-05-30 15:31:38.695783', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '40f748c7-3d13-4ef3-975b-066029ce4281', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (101, 5, 6, NULL, 'qsdjqsndnqs', '2025-05-30 15:31:39.432868', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'f27f2772-4788-4451-9470-587f410ffa4b', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (102, 5, 6, NULL, 'qs nddn', '2025-05-30 15:31:40.323567', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'cc14924f-00eb-4f70-93c0-f40bd61cf848', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (103, 5, 6, NULL, 'sqd,nd*sqdn', '2025-05-30 15:31:41.866816', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'd021fb49-de5c-4271-a9f4-8789ac41a6cc', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (104, 5, 6, NULL, 'qsndlqsd', '2025-05-30 15:31:42.921938', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '8f7d1de4-2cad-4c96-ac1d-8e0053b1f932', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (105, 5, 6, NULL, ',sqds', '2025-05-30 15:31:43.697512', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'd42ad729-79d2-43e0-bf43-3517e217e54c', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (106, 5, 6, NULL, ',sqds', '2025-05-30 15:31:44.533865', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '50b0755e-d8d9-47ab-b1d5-4bad440dd765', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (107, 5, 6, NULL, ',sqd', '2025-05-30 15:31:45.329471', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3437d38a-237f-4eaa-894f-331c17b62c03', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (108, 5, 6, NULL, 'sqd', '2025-05-30 15:31:46.198819', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3110a04e-b255-4128-a0ce-da569c7d3f62', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (109, 5, 6, NULL, 'qs,d', '2025-05-30 15:31:46.898809', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'eff02485-f547-467a-af04-fd4fde4eaab0', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (110, 5, 6, NULL, 'qsd,sq', '2025-05-30 15:31:47.617931', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '74b2f811-ad3e-450b-b486-e2b10dfbc29c', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (111, 5, 6, NULL, '*d,dqs', '2025-05-30 15:31:48.357704', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '49bb984c-a74f-4bfc-9bfa-4e036c73cf42', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (112, 5, 6, NULL, 'd,qs', '2025-05-30 15:31:49.040103', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'c1135891-248e-4f38-a5da-be4b87b31c83', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (113, 5, 6, NULL, 'sdq,dsq', '2025-05-30 15:31:49.742102', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '75d4dffe-0628-4f86-b198-fa92b6370511', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (120, 5, 7, NULL, 'test', '2025-06-03 13:03:35.018931', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '2c6f9db5-2fde-470b-a93a-d029252bde1c', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (72, 5, 6, NULL, 'ou à ballons ?', '2025-05-17 16:14:32.725111', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e3a3d5e7-810c-4233-9820-3564b2d5fede', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (73, 6, 5, NULL, 'à émotions ?', '2025-05-17 16:14:38.250447', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '59aa9c37-183b-492d-85a3-1fc9498230da', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (77, 7, 10, NULL, 'siouu', '2025-05-18 13:39:25.404866', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '8ccfc89d-c246-45c0-a607-596ae6c99815', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (78, 10, 7, NULL, 'ca va', '2025-05-18 13:39:28.713857', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '07006eb7-e4d7-4642-9f58-293ba5af448a', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (79, 7, 10, NULL, 'oui et toi', '2025-05-18 13:39:32.839204', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '9c0f9168-f8fe-4c88-9dd2-534137d4adae', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (80, 10, 7, NULL, 'ddzzz', '2025-05-18 15:04:13.479396', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '5b03f6f4-d4e4-4581-b038-884fa99e2e93', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (81, 10, 7, NULL, 'test', '2025-05-18 17:15:35.660566', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '0ddd9de7-9cf0-4924-bed0-107a5a9e1e70', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (82, 10, 7, NULL, 'test', '2025-05-18 17:18:39.995458', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '68cf65d0-8290-416d-862d-c22cc7f924d8', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (83, 10, 7, NULL, 'sympa', '2025-05-18 17:22:07.281441', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'd2454d32-9fd6-42b0-9b76-195b71d93356', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (89, 5, 10, NULL, 'test', '2025-05-30 14:57:39.556702', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '14247624-9a7e-44fd-9a5d-b7f0334881de', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (90, 5, 10, NULL, 'test avec marko', '2025-05-30 15:01:53.119959', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3ce62f68-7ed1-4cdb-a12a-82e8b329967f', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (91, 5, 10, NULL, 'markoo*', '2025-05-30 15:02:02.627174', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '819b06ee-3d31-4585-9ed4-d6d54b798c08', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (92, 5, 8, NULL, 'test avec marko', '2025-05-30 15:02:10.107562', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3efedfbd-d2aa-457b-94cc-86dad264799c', 3, NULL, '2025-06-18 23:56:00.769525+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (1, 3, NULL, 'general', 'testuser: Hello world', '2025-05-14 08:09:26.646215', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '34cb2aab-1b60-4d21-92bf-6600806379ba', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (2, 3, NULL, 'general', 'testuser: whouaaa c''est la zazou', '2025-05-14 10:04:01.46018', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b1856d34-da9c-4c53-9bf2-a490851eb381', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (3, 5, NULL, 'afterworks', 'test', '2025-05-16 17:58:48.605015', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'ff498c88-190f-4786-8e5c-f92b640b03a6', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (4, 5, NULL, 'general', 'test', '2025-05-16 17:59:28.625264', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'f33ae130-10c2-48fa-a36d-08959928ec40', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (5, 6, NULL, 'general', 'hey', '2025-05-16 18:01:01.85705', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1a7af8a3-1106-4dda-904c-087314aedeaa', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (6, 5, NULL, 'general', 'ca va ?', '2025-05-16 18:01:05.417042', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '082b9b52-0f62-4032-94a0-1a7b2a5aaa00', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (7, 6, NULL, 'general', 'oui et toi', '2025-05-16 18:01:10.620099', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'ab46e8c2-d339-48a9-b681-06288fccd082', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (8, 3, NULL, 'general', 'Hello from terminal!', '2025-05-17 12:28:20.87812', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '10f1ac51-2a20-4c64-ab5e-04430bd4212c', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (9, 6, NULL, 'general', 'siouu', '2025-05-17 13:07:18.683595', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '976d37db-663c-4972-a54b-8169f2fce406', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (10, 6, NULL, 'general', 'peka', '2025-05-17 13:07:21.773443', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '73a7c515-9880-4d5c-9420-43d66db095b1', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (11, 5, NULL, 'general', 'il y a un probleme ?', '2025-05-17 13:07:37.077497', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '00198652-4423-4bea-a3b9-e4690173705f', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (12, 6, NULL, 'general', 'siouu', '2025-05-17 13:07:41.390959', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '0596b224-5dca-48c6-8b82-c6fc85de7159', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (13, 6, NULL, 'general', 'test', '2025-05-17 13:22:49.022536', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e855f81c-03c8-4511-91b5-8acf654e7e24', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (14, 5, NULL, 'general', 'avion', '2025-05-17 13:22:52.362462', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1000e1d7-2fd8-4dcf-b959-0c3505cd2e88', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (15, 5, NULL, 'general', 'à réaction', '2025-05-17 13:22:56.760171', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1c0dcf97-f083-49dd-b74c-22a5aec96d2e', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (16, 5, NULL, 'general', 'dans la boue', '2025-05-17 13:23:00.717775', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'd8fa54d9-0a25-4d37-89e2-b33b73590cb9', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (17, 6, NULL, 'general', 'j''ai essayé', '2025-05-17 13:23:04.048055', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e2e69bb7-e2f9-4509-b075-0e114c68c1df', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (18, 5, NULL, 'afterworks', 'testing', '2025-05-17 13:34:24.56464', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'f2b68670-eea4-470f-846e-7221dbd6166c', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (19, 6, NULL, 'general', '<script>alert("hacked");*</script>', '2025-05-17 13:51:30.193538', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b9680171-fd7f-4449-9295-3b075cbdcf30', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (20, 6, NULL, 'general', 'tsty', '2025-05-17 14:05:44.174228', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '88efd9ba-9a54-4950-8c1c-e440907de565', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (21, 6, NULL, 'general', 'est', '2025-05-17 14:05:45.400753', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '6cbe4cb6-a032-4674-b234-ea18fec885cc', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (22, 6, NULL, 'general', 'fx', '2025-05-17 14:05:46.589543', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '2788115b-3855-4c3d-9123-c58f33a6b1c3', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (23, 5, NULL, 'general', 'testtt', '2025-05-17 14:06:12.653797', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '4c7bc1a7-6b2e-46ff-b6f2-20180530a15a', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (24, 5, NULL, 'general', 'esf', '2025-05-17 14:06:14.540518', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '40c94035-c7f2-4bc2-9d27-b950df5ebed4', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (25, 6, NULL, 'general', 'ok donc maintenant ca marche', '2025-05-17 14:13:57.097757', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '15de3de0-61a9-4989-b186-f57146ccfef3', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (26, 6, NULL, 'general', 'ya un probleme quand meme', '2025-05-17 14:14:03.587795', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1db472b1-2ed3-4327-88db-6ff2d0822887', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (27, 6, NULL, 'general', 'sqhqsj', '2025-05-17 14:16:56.810294', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b63df203-9bd5-480c-9c8d-556b36d21953', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (28, 6, NULL, 'general', 'en pétard', '2025-05-17 14:19:38.569998', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b2e40540-fce1-431f-8ce0-4fde6dbdb701', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (29, 5, NULL, 'general', 'ca riegole pas', '2025-05-17 14:19:50.480501', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e26eea13-bab0-4619-a780-167bd5887788', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (30, 5, NULL, 'general', 'c''est fou', '2025-05-17 14:19:54.539292', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'adb0f009-481a-4fef-9b7c-8407f8ba168a', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (31, 6, NULL, 'afterworks', 'incoyable', '2025-05-17 14:20:04.378719', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '08bd66a0-4421-4207-941b-579bb8758498', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (32, 5, NULL, 'afterworks', 'vraiment', '2025-05-17 14:20:09.634674', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'bb5fc062-3dce-4a2f-920a-0264d6235328', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (33, 5, NULL, 'general', 'genre la tout marche', '2025-05-17 14:20:15.205537', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1b124910-4a86-4aeb-ac8e-2d7676b39166', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (34, 5, NULL, 'general', 'c''est fou', '2025-05-17 14:20:32.662259', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '7c7b7022-ea14-4c6d-bdd8-c134806f30d9', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (35, 6, NULL, 'general', 'test incro', '2025-05-17 14:21:40.680769', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '03eeae9b-68c0-4c81-a57d-1869feb14dc7', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (36, 5, NULL, 'general', 'osqdjdd', '2025-05-17 14:21:44.353604', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3f70406b-c1eb-4c2f-b0f2-a73427b76102', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (37, 5, NULL, 'general', 'jsdkdd', '2025-05-17 14:21:48.11842', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '267af48c-7776-4a2c-bf36-d7deb4b43b9d', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (38, 5, NULL, 'general', 'fiouu', '2025-05-17 14:32:17.095472', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '7049c78e-8daf-4214-9f93-a2c930b5b5d5', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (39, 5, NULL, 'general', 'ajout', '2025-05-17 14:32:21.962477', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '99afa40c-8b11-42a2-b87a-a49c411b763b', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (40, 6, NULL, 'general', 'sjdl', '2025-05-17 14:32:24.519381', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b10cd4b8-82bf-4cca-85e2-1fd422de69ab', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (41, 6, NULL, 'afterworks', 'jdksdkd', '2025-05-17 14:32:27.897236', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '98937469-9535-4054-b003-69a4cc7622e7', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (42, 5, NULL, 'general', 'test', '2025-05-17 14:35:30.512052', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'fce4d849-0f3f-4d03-8f90-2ddefb0cf8c2', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (43, 5, NULL, 'general', 'ajout', '2025-05-17 14:35:32.153122', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '02c9fa63-1e15-4a45-9a2b-5073a3513f47', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (44, 6, NULL, 'general', 'final', '2025-05-17 14:35:35.082629', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'bad6cbfb-66fc-4478-9493-3619bc992425', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (45, 6, NULL, 'general', 'test', '2025-05-17 14:38:26.599281', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '630f0a38-78ba-4330-9ca7-a6abac266459', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (46, 5, NULL, 'general', 'avion', '2025-05-17 14:38:30.280868', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'f3b50cad-d633-40a5-ac76-8582e555b02b', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (47, 5, NULL, 'general', 'cachou', '2025-05-17 14:38:34.717786', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'ba9a82f9-f56f-4883-8080-3a726b383cf1', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (48, 5, NULL, 'general', 'test', '2025-05-17 14:38:41.694525', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'ab5a07b5-bde5-4486-a546-738956afdece', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (49, 6, NULL, 'afterworks', 'callera', '2025-05-17 14:38:46.590603', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '869460cc-43d6-494a-adba-a67894d8ceae', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (50, 5, NULL, 'afterworks', 'test', '2025-05-17 14:38:52.859089', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '42629dff-bda2-469e-b338-c0779d4fcfe5', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (51, 6, NULL, 'general', 'jss', '2025-05-17 14:44:14.183201', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '765ccb6f-239b-4306-a638-d5d625cae6d9', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (52, 6, NULL, 'general', 'testing', '2025-05-17 15:05:53.881917', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'bb70205d-77a0-47b0-b05e-0bf03501051b', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (53, 5, NULL, 'general', 'seems to work', '2025-05-17 15:06:05.664446', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'c9481beb-29b3-45b2-91a3-1bf8c12ccd08', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (54, 5, NULL, 'general', 'we''ll see', '2025-05-17 15:06:09.661665', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '6b217f1a-acb3-46fe-a3c0-32a06ee47aab', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (55, 5, NULL, 'general', 'toto is now in afterworks room', '2025-05-17 15:06:20.931807', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '7d5e465e-54a7-408f-b0b7-957258b794b8', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (56, 6, NULL, 'afterworks', 'he doesn''t see a zoukou msg from egneral room', '2025-05-17 15:06:33.487061', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e3f98726-ec6c-4995-80a0-77285c9f6107', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (57, 6, NULL, 'afterworks', 'so bug fixed', '2025-05-17 15:06:35.895411', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '7a8feedd-82d9-4c00-9909-8bfcafa96af3', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (58, 6, NULL, 'general', 'fou', '2025-05-17 15:10:18.93665', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '3bfbe811-c5e1-4359-9921-1a54fdc088a1', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (59, 6, NULL, 'afterworks', 'incr', '2025-05-17 15:10:31.571454', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '256ab3b7-518a-4606-8a98-84366aeda019', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (60, 5, NULL, 'general', 'fifi', '2025-05-17 15:20:40.624171', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'a4cdf12c-8cfc-4c10-a671-89753b32e935', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (74, 7, NULL, 'general', 'coucou', '2025-05-18 13:01:57.684899', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '7786061b-58c1-4b54-92a7-3b7253664fb9', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (75, 7, NULL, 'general', 'coucou', '2025-05-18 13:02:09.652751', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '97284d03-a4d9-46b4-aedb-f16b26f969bf', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (76, 10, NULL, 'general', 'coucou', '2025-05-18 13:02:24.060334', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '5f8e1b71-e834-44c3-a22c-333d10cd7f49', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (84, 5, NULL, 'general', 'sifiliiii', '2025-05-30 14:28:15.497466', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b480ce1b-97c3-4f0e-8477-9ac2db253ed6', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (85, 5, NULL, 'afterworks', 'dodo', '2025-05-30 14:28:21.313166', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '784292a8-e22c-4ba2-b274-4f9b2428cf84', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (86, 5, NULL, 'afterworks', 'sifili', '2025-05-30 14:28:30.906011', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '8b326f19-bd32-4d9c-9c2a-60c446e165cc', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (87, 5, NULL, 'general', 'sifili', '2025-05-30 14:28:38.138026', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '001af8b6-f139-4508-ba42-834c79a85488', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (88, 5, NULL, 'general', 'sifili', '2025-05-30 14:28:51.67819', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'c560bfb2-84ac-4b76-a0cb-15e9bb9e2ca3', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (114, 12, NULL, 'afterworks', 'castel red', '2025-05-31 15:46:00.919832', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'ffc2f220-1b16-4b09-90b5-072585ea308f', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (115, 12, NULL, 'afterworks', 'nikola', '2025-05-31 15:46:11.339157', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'e3084709-89da-43c1-8dd7-19eecd3a4a50', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (116, 12, NULL, 'afterworks', 'oui oui', '2025-06-03 12:41:07.267214', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '1c16cda8-5040-4700-816b-2e94224bf4e1', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (117, 12, NULL, 'afterworks', 'test', '2025-06-03 12:44:03.142528', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '20c30077-5167-4031-911f-8655d153a687', 1, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (118, 12, NULL, 'general', 'test', '2025-06-03 12:44:08.136559', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'b804cfc4-1420-4831-838c-8394f74b63a4', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (119, 12, NULL, 'general', 'fouuu', '2025-06-03 12:44:11.166248', 'text', NULL, false, NULL, NULL, false, 0, 'sent', '55b80b7c-356a-44ee-989c-02635c92b160', 2, NULL, '2025-06-18 23:56:00.714678+00');
INSERT INTO "public"."messages" ("id", "author_id", "to_user", "room", "content", "created_at", "message_type", "reply_to_id", "is_edited", "edited_at", "metadata", "is_pinned", "thread_count", "status", "uuid", "conversation_id", "parent_message_id", "updated_at") VALUES (121, 5, NULL, 'afterworks', 'zoo', '2025-06-03 13:19:04.487317', 'text', NULL, false, NULL, NULL, false, 0, 'sent', 'acfdd231-4d54-4f4e-9223-07bc950ffe41', 1, NULL, '2025-06-18 23:56:00.714678+00');


--
-- TOC entry 3778 (class 0 OID 24885)
-- Dependencies: 238
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."migrations" ("id", "filename", "applied_at") VALUES (1, '001_users.sql', '2025-06-04 16:58:28.04282');
INSERT INTO "public"."migrations" ("id", "filename", "applied_at") VALUES (2, 'files.sql', '2025-06-04 18:22:46.819163');
INSERT INTO "public"."migrations" ("id", "filename", "applied_at") VALUES (3, 'internal_ressources.sql', '2025-06-04 18:22:46.822494');


--
-- TOC entry 3792 (class 0 OID 25078)
-- Dependencies: 252
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3772 (class 0 OID 24729)
-- Dependencies: 232
-- Data for Name: offers; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- TOC entry 3776 (class 0 OID 24841)
-- Dependencies: 236
-- Data for Name: product_documents; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3774 (class 0 OID 24771)
-- Dependencies: 234
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (1, 'Microphone USB', 1, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.916989');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (2, 'Interface audio', 2, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.918816');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (3, 'Casque monitoring', 3, '', '', '', NULL, 12, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.919537');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (4, 'Enceintes de studio', 4, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.920173');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (5, 'Contrôleur MIDI', 5, '', '', '', NULL, 12, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.920727');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (6, 'Table de mixage', 6, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.921408');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (7, 'Préampli micro', 7, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.92195');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (8, 'Compresseur', 7, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.92195');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (9, 'Égaliseur', 7, '', '', '', NULL, 24, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.92195');
INSERT INTO "public"."products" ("id", "name", "category_id", "brand", "model", "description", "price", "warranty_months", "warranty_conditions", "manufacturer_website", "specifications", "status", "created_at", "updated_at") VALUES (10, 'Câble XLR', 8, '', '', '', NULL, 6, '', '', '', 'active', '2025-06-04 10:51:30.914438', '2025-06-04 10:51:30.925163');


--
-- TOC entry 3782 (class 0 OID 24956)
-- Dependencies: 242
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (1, 6, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2LCJ1c2VybmFtZSI6InRlc3RlciIsInJvbGUiOiJ1c2VyIiwiZXhwIjoxNzUwMTgyMzAxLCJpYXQiOjE3NDk1Nzc1MDF9.5itXtCtghbj5z4eallR2oe-pWcG1D30TFKkGQZpcv7k', '2025-06-17 17:45:01.919772', '2025-06-10 17:45:01.919772');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (5, 7, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo3LCJ1c2VybmFtZSI6ImZpbG91Iiwicm9sZSI6InVzZXIiLCJleHAiOjE3NTAyMzUxOTMsImlhdCI6MTc0OTYzMDM5M30.3luCJtwfF45mkm1Xu0TZQZFHZfOlOY1DRTNa_c1TlXE', '2025-06-18 08:26:33.976361', '2025-06-11 08:26:33.976361');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (17, 9, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo5LCJ1c2VybmFtZSI6ImF2aW9uIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NTAyMzU5NTgsImlhdCI6MTc0OTYzMTE1OH0.ZuaPisUgLVpRAhHKyhKYmbYpC-_7UIgyrCgvGQdKUpw', '2025-06-18 08:39:18.979907', '2025-06-11 08:39:18.979907');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (20, 10, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMCwidXNlcm5hbWUiOiJuaWtvIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NTAyNjgwNDgsImlhdCI6MTc0OTY2MzI0OH0.x9qzymi4Q12LjojOB28XK_M-vUN_RkWchOZbQ8muCYg', '2025-06-18 17:34:08.661378', '2025-06-11 17:34:08.661378');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (22, 11, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMSwidXNlcm5hbWUiOiJ0ZXN0Iiwicm9sZSI6InVzZXIiLCJleHAiOjE3NTAzNTYyNTUsImlhdCI6MTc0OTc1MTQ1NX0.8vwhM9C9ODKxmqPVPFh3ES_KBxk5N57MtKLrsb9ZX40', '2025-06-19 18:04:15.094627', '2025-06-12 18:04:15.094627');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (43, 15, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNSwidXNlcm5hbWUiOiJ0ZXN0Y2hhdCIsInJvbGUiOiJ1c2VyIiwiZXhwIjoxNzUwNjY2MzQ5LCJpYXQiOjE3NTAwNjE1NDl9.bbHwXplbwVeBMh4loJKc4Py7xqP_LgaIAsyS8xMaSew', '2025-06-23 08:12:29.181905', '2025-06-16 08:12:29.181905');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (55, 14, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxNCwidXNlcm5hbWUiOiJsb3Vsb3UiLCJyb2xlIjoidXNlciIsImV4cCI6MTc1MTA0NzAyNSwiaWF0IjoxNzUwNDQyMjI1fQ.LQvPbaYV7r_aLcKVKLq8P4LIM5Ubv3qPH2W8xnaHKhk', '2025-06-27 17:57:05.93605', '2025-06-20 17:57:05.93605');
INSERT INTO "public"."refresh_tokens" ("id", "user_id", "token", "expires_at", "created_at") VALUES (16, 8, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo4LCJ1c2VybmFtZSI6ImtvdWJvdSIsInJvbGUiOiJ1c2VyIiwiZXhwIjoxNzUxMDQ3MTExLCJpYXQiOjE3NTA0NDIzMTF9.NXA3rozXI_OYNKEL3eiJqhcv6ZU0bQ5cSMJQdztqEnk', '2025-06-27 17:58:31.46907', '2025-06-20 17:58:31.46907');


--
-- TOC entry 3790 (class 0 OID 25054)
-- Dependencies: 250
-- Data for Name: room_members; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3766 (class 0 OID 16466)
-- Dependencies: 226
-- Data for Name: rooms; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."rooms" ("id", "name", "is_private", "created_at", "creator_id", "max_members", "description") VALUES (1, 'general', false, '2025-05-14 08:05:03.902599', NULL, 1000, NULL);
INSERT INTO "public"."rooms" ("id", "name", "is_private", "created_at", "creator_id", "max_members", "description") VALUES (2, 'afterworks', false, '2025-05-14 08:06:14.174808', NULL, 1000, NULL);


--
-- TOC entry 3784 (class 0 OID 24985)
-- Dependencies: 244
-- Data for Name: sanctions; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3770 (class 0 OID 16550)
-- Dependencies: 230
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO "public"."tags" ("id", "name") VALUES (1, 'hiphop');
INSERT INTO "public"."tags" ("id", "name") VALUES (2, 'ambient');
INSERT INTO "public"."tags" ("id", "name") VALUES (3, 'techno');
INSERT INTO "public"."tags" ("id", "name") VALUES (4, 'trap');
INSERT INTO "public"."tags" ("id", "name") VALUES (5, 'lofi');
INSERT INTO "public"."tags" ("id", "name") VALUES (6, 'synth');


--
-- TOC entry 3788 (class 0 OID 25029)
-- Dependencies: 248
-- Data for Name: user_blocks; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3794 (class 0 OID 25097)
-- Dependencies: 254
-- Data for Name: user_sessions; Type: TABLE DATA; Schema: public; Owner: veza
--



--
-- TOC entry 3780 (class 0 OID 24924)
-- Dependencies: 240
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: veza
--

INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (1, 'test_user', 'test@free.fr', '$2a$10$zqSerwfpsErKDKB/s3rnYuDBCs9AkwntWFUTrTBV3xDhCLDCcYWQq', '', '', '', '', 'user', true, false, NULL, '2025-06-06 15:53:31.838708', '2025-06-06 15:53:31.838708', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '2871d7a4-0d50-429c-b0ea-1d96475f3877', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (2, 'testuser', 'test@example.com', '$2a$10$1CBH45rwl3OXdNHH9Sgt9O.MmHHvxjpd5uZIm9FT5lccRHB1HlRQC', '', '', '', '', 'user', true, false, NULL, '2025-06-06 16:02:08.821544', '2025-06-06 16:02:08.821544', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'b0b4d429-003d-4408-be0c-450851f4b3ac', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (3, 'newuser', 'new@example.com', '$2a$10$TcUbD6arex3jgPUeEZ1RPOYdvNdIfr/vpwtEedlhP2iPMtkp8jHw6', '', '', '', '', 'user', true, false, NULL, '2025-06-06 16:16:11.234989', '2025-06-06 16:16:11.234989', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'f90b8e2f-00e0-4e7b-bc4a-5bd3f293e667', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (4, 'newuser2', 'new2@example.com', '$2a$10$gPFVka9twToCAb3CMGaCPOXLHjzzMqaocun77d2iHP4o0N/zkqbTS', '', '', '', '', 'user', true, false, NULL, '2025-06-06 16:31:54.479521', '2025-06-06 16:31:54.479521', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '414cf533-8b7a-4b66-8195-39e7990efa52', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (5, 'harry', 'harry@example.com', '$2a$10$6WHrKaqVzsiD5A.yZAMjFOtaCPJ1MIuI.d25FjxRdsJFTOM28YwJC', '', '', '', '', 'user', true, false, NULL, '2025-06-06 17:33:55.844063', '2025-06-06 17:33:55.844063', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '976a8c18-d0ba-440e-ba15-75a310a68369', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (6, 'tester', 'tester@free.fr', '$2a$10$Tukex8wH0iLa40Bh.qld6eEf43GWPY.xC0WuKUedafkEtUmlzxmgy', '', '', '', '', 'user', true, false, NULL, '2025-06-09 15:05:26.840141', '2025-06-09 15:05:26.840141', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '040a17b6-334a-4b48-b7cb-f76313df7222', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (7, 'filou', 'filou@example.com', '$2a$10$sy/ZsrHNATYbraeUaeGjq.m3sDwomn1SlClVgrO3meCrJ53Ng6g.a', '', '', '', '', 'user', true, false, NULL, '2025-06-09 16:47:26.914866', '2025-06-09 16:47:26.914866', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '21994a59-f2bb-4b04-8516-6ef24a607b5a', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (8, 'koubou', 'koubou@example.com', '$2a$10$lu3WGzaAK73pUUM1cyAGWeYFA2kIpx/5rBtyquXgpnEL4I6clbTee', '', '', '', '', 'user', true, false, NULL, '2025-06-11 08:28:13.931282', '2025-06-11 08:28:13.931282', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '65cb5796-abfb-42e3-b152-d3364c797eb2', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (9, 'avion', 'avion@free.fr', '$2a$10$DaPn4lEhppsvvxM5xi2LauDP3YcfaE6cMeMR9cfo4bx9nLV7fdKcW', '', '', '', '', 'user', true, false, NULL, '2025-06-11 08:35:59.612549', '2025-06-11 08:35:59.612549', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'd9efa9fe-a50a-4b3e-a052-a550c1898de4', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (10, 'niko', 'niko@free.fr', '$2a$10$8iVau2eAxznWlm0XlSvYVei3Z8lg3P0lK6dYi9ryRXzX8b2TSEpvC', '', '', '', '', 'user', true, false, NULL, '2025-06-11 17:33:58.607696', '2025-06-11 17:33:58.607696', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'f28f916e-9ea3-495b-b7b9-fdb0db628b5c', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (11, 'test', 'test@a.fr', '$2a$10$PNlt1NxLwq6W3ukIbSZ2ve1gkpuFonLli5oZsf55XF2TAa/Z0k/8K', '', '', '', '', 'user', true, false, NULL, '2025-06-12 17:56:36.371586', '2025-06-12 17:56:36.371586', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '1be9560c-04f7-4f45-b2c8-a2f7d51a7afa', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (12, 'shelby', 'shelby@free.fr', '$2a$10$N6z8ZdHvUTqiT6Ugq2iiYuECnm6Mb2ebiBcZgMM30VJMUoJBMZ07G', '', '', '', '', 'user', true, false, NULL, '2025-06-13 20:44:47.878159', '2025-06-13 20:44:47.878159', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'd9d6e96f-c86f-4181-a321-fe9f42e14eda', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (13, 'panda', 'panda@free.fr', '$2a$10$tkDptfrSpJ.PgUjfHcKaXOTypxeUsLG9B7gzlHCgHKZ/LB2WRfory', '', '', '', '', 'user', true, false, NULL, '2025-06-13 20:46:47.469178', '2025-06-13 20:46:47.469178', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '4b085774-55a6-4cdb-b1c7-995fbd3f5b8a', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (14, 'loulou', 'loulou@free.fr', '$2a$10$1/Y0oSNjbMhPdE.lgrBD9uLC91lgE8tQfomEvUS2sMNZtAtsIvCO.', '', '', '', '', 'user', true, false, NULL, '2025-06-13 21:03:54.702786', '2025-06-13 21:03:54.702786', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, 'c6e660c2-f6b1-4cce-93cb-589586026e29', '2025-06-18 23:56:00.679051+00');
INSERT INTO "public"."users" ("id", "username", "email", "password_hash", "first_name", "last_name", "avatar", "bio", "role", "is_active", "is_verified", "last_login_at", "created_at", "updated_at", "status", "last_seen", "reputation_score", "is_banned", "is_muted", "uuid", "last_activity") VALUES (15, 'testchat', 'testchat@example.com', '$2a$10$rrc2jDHBcGJT7V/3OXMH.elkEXuDmK1IGsy9TbexKQT..VK3eso3a', '', '', '', '', 'user', true, false, NULL, '2025-06-16 01:11:11.817287', '2025-06-16 01:11:11.817287', 'offline', '2025-06-18 21:53:32.702861+00', 100, false, false, '10e28b35-f11b-4e10-9961-5f762d3af18f', '2025-06-18 23:56:00.679051+00');


--
-- TOC entry 3848 (class 0 OID 0)
-- Dependencies: 255
-- Name: audit_logs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."audit_logs_id_seq"', 1, false);


--
-- TOC entry 3849 (class 0 OID 0)
-- Dependencies: 259
-- Name: conversation_members_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."conversation_members_id_seq"', 1, false);


--
-- TOC entry 3850 (class 0 OID 0)
-- Dependencies: 257
-- Name: conversations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."conversations_id_seq"', 3, true);


--
-- TOC entry 3851 (class 0 OID 0)
-- Dependencies: 223
-- Name: files_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."files_id_seq"', 1, true);


--
-- TOC entry 3852 (class 0 OID 0)
-- Dependencies: 263
-- Name: message_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."message_history_id_seq"', 1, false);


--
-- TOC entry 3853 (class 0 OID 0)
-- Dependencies: 261
-- Name: message_mentions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."message_mentions_id_seq"', 1, false);


--
-- TOC entry 3854 (class 0 OID 0)
-- Dependencies: 245
-- Name: message_reactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."message_reactions_id_seq"', 1, false);


--
-- TOC entry 3855 (class 0 OID 0)
-- Dependencies: 227
-- Name: messages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."messages_id_seq"', 126, true);


--
-- TOC entry 3856 (class 0 OID 0)
-- Dependencies: 237
-- Name: migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."migrations_id_seq"', 3, true);


--
-- TOC entry 3857 (class 0 OID 0)
-- Dependencies: 251
-- Name: notifications_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."notifications_id_seq"', 1, false);


--
-- TOC entry 3858 (class 0 OID 0)
-- Dependencies: 231
-- Name: offers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('"public"."offers_id_seq"', 5, true);


--
-- TOC entry 3859 (class 0 OID 0)
-- Dependencies: 235
-- Name: product_documents_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."product_documents_id_seq"', 1, false);


--
-- TOC entry 3860 (class 0 OID 0)
-- Dependencies: 233
-- Name: products_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('"public"."products_id_seq1"', 21, true);


--
-- TOC entry 3861 (class 0 OID 0)
-- Dependencies: 241
-- Name: refresh_tokens_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('"public"."refresh_tokens_id_seq"', 56, true);


--
-- TOC entry 3862 (class 0 OID 0)
-- Dependencies: 249
-- Name: room_members_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."room_members_id_seq"', 1, false);


--
-- TOC entry 3863 (class 0 OID 0)
-- Dependencies: 225
-- Name: rooms_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."rooms_id_seq"', 2, true);


--
-- TOC entry 3864 (class 0 OID 0)
-- Dependencies: 243
-- Name: sanctions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."sanctions_id_seq"', 1, false);


--
-- TOC entry 3865 (class 0 OID 0)
-- Dependencies: 229
-- Name: tags_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('"public"."tags_id_seq"', 6, true);


--
-- TOC entry 3866 (class 0 OID 0)
-- Dependencies: 247
-- Name: user_blocks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."user_blocks_id_seq"', 1, false);


--
-- TOC entry 3867 (class 0 OID 0)
-- Dependencies: 253
-- Name: user_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."user_sessions_id_seq"', 1, false);


--
-- TOC entry 3868 (class 0 OID 0)
-- Dependencies: 239
-- Name: users_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: veza
--

SELECT pg_catalog.setval('"public"."users_id_seq1"', 15, true);


--
-- TOC entry 3580 (class 2606 OID 25806)
-- Name: conversation_members conversation_members_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversation_members"
    ADD CONSTRAINT "conversation_members_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3582 (class 2606 OID 25808)
-- Name: conversation_members conversation_members_unique; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversation_members"
    ADD CONSTRAINT "conversation_members_unique" UNIQUE ("conversation_id", "user_id");


--
-- TOC entry 3574 (class 2606 OID 25771)
-- Name: conversations conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversations"
    ADD CONSTRAINT "conversations_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3576 (class 2606 OID 25773)
-- Name: conversations conversations_uuid_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversations"
    ADD CONSTRAINT "conversations_uuid_key" UNIQUE ("uuid");


--
-- TOC entry 3474 (class 2606 OID 16444)
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."files"
    ADD CONSTRAINT "files_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3589 (class 2606 OID 25849)
-- Name: message_history message_history_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_history"
    ADD CONSTRAINT "message_history_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3585 (class 2606 OID 25827)
-- Name: message_mentions message_mentions_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_mentions"
    ADD CONSTRAINT "message_mentions_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3587 (class 2606 OID 25829)
-- Name: message_mentions message_mentions_unique; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_mentions"
    ADD CONSTRAINT "message_mentions_unique" UNIQUE ("message_id", "mentioned_user_id");


--
-- TOC entry 3538 (class 2606 OID 25016)
-- Name: message_reactions message_reactions_message_id_user_id_reaction_type_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_reactions"
    ADD CONSTRAINT "message_reactions_message_id_user_id_reaction_type_key" UNIQUE ("message_id", "user_id", "reaction_type");


--
-- TOC entry 3540 (class 2606 OID 25014)
-- Name: message_reactions message_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_reactions"
    ADD CONSTRAINT "message_reactions_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3488 (class 2606 OID 16487)
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3490 (class 2606 OID 25791)
-- Name: messages messages_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_uuid_unique" UNIQUE ("uuid");


--
-- TOC entry 3509 (class 2606 OID 24893)
-- Name: migrations migrations_filename_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."migrations"
    ADD CONSTRAINT "migrations_filename_key" UNIQUE ("filename");


--
-- TOC entry 3511 (class 2606 OID 24891)
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."migrations"
    ADD CONSTRAINT "migrations_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3558 (class 2606 OID 25087)
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."notifications"
    ADD CONSTRAINT "notifications_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3496 (class 2606 OID 24738)
-- Name: offers offers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."offers"
    ADD CONSTRAINT "offers_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3507 (class 2606 OID 24852)
-- Name: product_documents product_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."product_documents"
    ADD CONSTRAINT "product_documents_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3501 (class 2606 OID 24780)
-- Name: products products_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."products"
    ADD CONSTRAINT "products_name_key" UNIQUE ("name");


--
-- TOC entry 3503 (class 2606 OID 24778)
-- Name: products products_pkey1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."products"
    ADD CONSTRAINT "products_pkey1" PRIMARY KEY ("id");


--
-- TOC entry 3526 (class 2606 OID 24964)
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."refresh_tokens"
    ADD CONSTRAINT "refresh_tokens_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3528 (class 2606 OID 24966)
-- Name: refresh_tokens refresh_tokens_token_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."refresh_tokens"
    ADD CONSTRAINT "refresh_tokens_token_key" UNIQUE ("token");


--
-- TOC entry 3530 (class 2606 OID 24968)
-- Name: refresh_tokens refresh_tokens_user_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."refresh_tokens"
    ADD CONSTRAINT "refresh_tokens_user_id_key" UNIQUE ("user_id");


--
-- TOC entry 3551 (class 2606 OID 25061)
-- Name: room_members room_members_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."room_members"
    ADD CONSTRAINT "room_members_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3553 (class 2606 OID 25063)
-- Name: room_members room_members_room_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."room_members"
    ADD CONSTRAINT "room_members_room_id_user_id_key" UNIQUE ("room_id", "user_id");


--
-- TOC entry 3478 (class 2606 OID 16477)
-- Name: rooms rooms_name_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."rooms"
    ADD CONSTRAINT "rooms_name_key" UNIQUE ("name");


--
-- TOC entry 3480 (class 2606 OID 16475)
-- Name: rooms rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."rooms"
    ADD CONSTRAINT "rooms_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3534 (class 2606 OID 24994)
-- Name: sanctions sanctions_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."sanctions"
    ADD CONSTRAINT "sanctions_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3492 (class 2606 OID 16559)
-- Name: tags tags_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."tags"
    ADD CONSTRAINT "tags_name_key" UNIQUE ("name");


--
-- TOC entry 3494 (class 2606 OID 16557)
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."tags"
    ADD CONSTRAINT "tags_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3544 (class 2606 OID 25038)
-- Name: user_blocks user_blocks_blocker_id_blocked_id_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_blocks"
    ADD CONSTRAINT "user_blocks_blocker_id_blocked_id_key" UNIQUE ("blocker_id", "blocked_id");


--
-- TOC entry 3546 (class 2606 OID 25036)
-- Name: user_blocks user_blocks_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_blocks"
    ADD CONSTRAINT "user_blocks_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3566 (class 2606 OID 25107)
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_sessions"
    ADD CONSTRAINT "user_sessions_pkey" PRIMARY KEY ("id");


--
-- TOC entry 3568 (class 2606 OID 25109)
-- Name: user_sessions user_sessions_session_token_key; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_sessions"
    ADD CONSTRAINT "user_sessions_session_token_key" UNIQUE ("session_token");


--
-- TOC entry 3518 (class 2606 OID 24945)
-- Name: users users_email_key1; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_email_key1" UNIQUE ("email");


--
-- TOC entry 3520 (class 2606 OID 24941)
-- Name: users users_pkey1; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_pkey1" PRIMARY KEY ("id");


--
-- TOC entry 3522 (class 2606 OID 24943)
-- Name: users users_username_key1; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_username_key1" UNIQUE ("username");


--
-- TOC entry 3524 (class 2606 OID 25754)
-- Name: users users_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_uuid_unique" UNIQUE ("uuid");


--
-- TOC entry 3569 (class 1259 OID 25135)
-- Name: idx_audit_logs_action; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_audit_logs_action" ON "public"."audit_logs" USING "btree" ("action");


--
-- TOC entry 3570 (class 1259 OID 25137)
-- Name: idx_audit_logs_created; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_audit_logs_created" ON "public"."audit_logs" USING "btree" ("created_at");


--
-- TOC entry 3571 (class 1259 OID 25136)
-- Name: idx_audit_logs_resource; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_audit_logs_resource" ON "public"."audit_logs" USING "btree" ("resource_type", "resource_id");


--
-- TOC entry 3572 (class 1259 OID 25134)
-- Name: idx_audit_logs_user; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_audit_logs_user" ON "public"."audit_logs" USING "btree" ("user_id");


--
-- TOC entry 3577 (class 1259 OID 25863)
-- Name: idx_conversations_owner_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_conversations_owner_active" ON "public"."conversations" USING "btree" ("owner_id") WHERE (NOT "is_archived");


--
-- TOC entry 3578 (class 1259 OID 25862)
-- Name: idx_conversations_type_public; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_conversations_type_public" ON "public"."conversations" USING "btree" ("type") WHERE ("is_public" = true);


--
-- TOC entry 3583 (class 1259 OID 25867)
-- Name: idx_mentions_user_unread; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_mentions_user_unread" ON "public"."message_mentions" USING "btree" ("mentioned_user_id") WHERE ("is_read" = false);


--
-- TOC entry 3535 (class 1259 OID 25027)
-- Name: idx_message_reactions_message; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_message_reactions_message" ON "public"."message_reactions" USING "btree" ("message_id");


--
-- TOC entry 3481 (class 1259 OID 25865)
-- Name: idx_messages_author_time; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_author_time" ON "public"."messages" USING "btree" ("author_id", "created_at" DESC);


--
-- TOC entry 3482 (class 1259 OID 25864)
-- Name: idx_messages_conversation_time; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_conversation_time" ON "public"."messages" USING "btree" ("conversation_id", "created_at" DESC);


--
-- TOC entry 3483 (class 1259 OID 25147)
-- Name: idx_messages_edited; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_edited" ON "public"."messages" USING "btree" ("is_edited") WHERE ("is_edited" = true);


--
-- TOC entry 3484 (class 1259 OID 25420)
-- Name: idx_messages_pinned; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_pinned" ON "public"."messages" USING "btree" ("is_pinned") WHERE ("is_pinned" = true);


--
-- TOC entry 3485 (class 1259 OID 25146)
-- Name: idx_messages_reply; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_reply" ON "public"."messages" USING "btree" ("reply_to_id");


--
-- TOC entry 3486 (class 1259 OID 25145)
-- Name: idx_messages_type; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_messages_type" ON "public"."messages" USING "btree" ("message_type");


--
-- TOC entry 3554 (class 1259 OID 25095)
-- Name: idx_notifications_type; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_notifications_type" ON "public"."notifications" USING "btree" ("type");


--
-- TOC entry 3555 (class 1259 OID 25094)
-- Name: idx_notifications_unread; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_notifications_unread" ON "public"."notifications" USING "btree" ("user_id", "is_read") WHERE ("is_read" = false);


--
-- TOC entry 3556 (class 1259 OID 25093)
-- Name: idx_notifications_user; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_notifications_user" ON "public"."notifications" USING "btree" ("user_id");


--
-- TOC entry 3504 (class 1259 OID 24859)
-- Name: idx_product_documents_file_type; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_product_documents_file_type" ON "public"."product_documents" USING "btree" ("file_type");


--
-- TOC entry 3505 (class 1259 OID 24858)
-- Name: idx_product_documents_product_id; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_product_documents_product_id" ON "public"."product_documents" USING "btree" ("product_id");


--
-- TOC entry 3497 (class 1259 OID 24879)
-- Name: idx_products_category_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "idx_products_category_id" ON "public"."products" USING "btree" ("category_id");


--
-- TOC entry 3498 (class 1259 OID 24880)
-- Name: idx_products_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "idx_products_status" ON "public"."products" USING "btree" ("status");


--
-- TOC entry 3499 (class 1259 OID 24881)
-- Name: idx_products_updated_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX "idx_products_updated_at" ON "public"."products" USING "btree" ("updated_at");


--
-- TOC entry 3536 (class 1259 OID 25866)
-- Name: idx_reactions_user; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_reactions_user" ON "public"."message_reactions" USING "btree" ("user_id", "created_at" DESC);


--
-- TOC entry 3547 (class 1259 OID 25076)
-- Name: idx_room_members_role; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_room_members_role" ON "public"."room_members" USING "btree" ("room_id", "role");


--
-- TOC entry 3548 (class 1259 OID 25074)
-- Name: idx_room_members_room; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_room_members_room" ON "public"."room_members" USING "btree" ("room_id");


--
-- TOC entry 3549 (class 1259 OID 25075)
-- Name: idx_room_members_user; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_room_members_user" ON "public"."room_members" USING "btree" ("user_id");


--
-- TOC entry 3475 (class 1259 OID 25051)
-- Name: idx_rooms_name; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_rooms_name" ON "public"."rooms" USING "btree" ("name");


--
-- TOC entry 3476 (class 1259 OID 25052)
-- Name: idx_rooms_private; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_rooms_private" ON "public"."rooms" USING "btree" ("is_private");


--
-- TOC entry 3531 (class 1259 OID 25006)
-- Name: idx_sanctions_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_sanctions_active" ON "public"."sanctions" USING "btree" ("user_id", "is_active") WHERE ("is_active" = true);


--
-- TOC entry 3532 (class 1259 OID 25005)
-- Name: idx_sanctions_user_id; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_sanctions_user_id" ON "public"."sanctions" USING "btree" ("user_id");


--
-- TOC entry 3559 (class 1259 OID 25869)
-- Name: idx_sessions_token; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_sessions_token" ON "public"."user_sessions" USING "btree" ("session_token");


--
-- TOC entry 3560 (class 1259 OID 25868)
-- Name: idx_sessions_user_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_sessions_user_active" ON "public"."user_sessions" USING "btree" ("user_id") WHERE ("is_active" = true);


--
-- TOC entry 3541 (class 1259 OID 25050)
-- Name: idx_user_blocks_blocked; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_blocks_blocked" ON "public"."user_blocks" USING "btree" ("blocked_id");


--
-- TOC entry 3542 (class 1259 OID 25049)
-- Name: idx_user_blocks_blocker; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_blocks_blocker" ON "public"."user_blocks" USING "btree" ("blocker_id");


--
-- TOC entry 3561 (class 1259 OID 25117)
-- Name: idx_user_sessions_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_sessions_active" ON "public"."user_sessions" USING "btree" ("user_id", "is_active") WHERE ("is_active" = true);


--
-- TOC entry 3562 (class 1259 OID 25118)
-- Name: idx_user_sessions_expires; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_sessions_expires" ON "public"."user_sessions" USING "btree" ("expires_at");


--
-- TOC entry 3563 (class 1259 OID 25116)
-- Name: idx_user_sessions_token; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_sessions_token" ON "public"."user_sessions" USING "btree" ("session_token");


--
-- TOC entry 3564 (class 1259 OID 25115)
-- Name: idx_user_sessions_user; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_user_sessions_user" ON "public"."user_sessions" USING "btree" ("user_id");


--
-- TOC entry 3512 (class 1259 OID 24948)
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_users_created_at" ON "public"."users" USING "btree" ("created_at");


--
-- TOC entry 3513 (class 1259 OID 25861)
-- Name: idx_users_email_verified; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_users_email_verified" ON "public"."users" USING "btree" ("email") WHERE ("is_verified" = true);


--
-- TOC entry 3514 (class 1259 OID 24947)
-- Name: idx_users_is_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_users_is_active" ON "public"."users" USING "btree" ("is_active");


--
-- TOC entry 3515 (class 1259 OID 25153)
-- Name: idx_users_status; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_users_status" ON "public"."users" USING "btree" ("status");


--
-- TOC entry 3516 (class 1259 OID 25860)
-- Name: idx_users_username_active; Type: INDEX; Schema: public; Owner: veza
--

CREATE INDEX "idx_users_username_active" ON "public"."users" USING "btree" ("username") WHERE ("is_active" = true);


--
-- TOC entry 3617 (class 2620 OID 25890)
-- Name: conversations update_conversations_updated_at; Type: TRIGGER; Schema: public; Owner: veza
--

CREATE TRIGGER "update_conversations_updated_at" BEFORE UPDATE ON "public"."conversations" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();


--
-- TOC entry 3614 (class 2620 OID 25891)
-- Name: messages update_messages_updated_at; Type: TRIGGER; Schema: public; Owner: veza
--

CREATE TRIGGER "update_messages_updated_at" BEFORE UPDATE ON "public"."messages" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();


--
-- TOC entry 3615 (class 2620 OID 24882)
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER "update_products_updated_at" BEFORE UPDATE ON "public"."products" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();


--
-- TOC entry 3616 (class 2620 OID 25889)
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: veza
--

CREATE TRIGGER "update_users_updated_at" BEFORE UPDATE ON "public"."users" FOR EACH ROW EXECUTE FUNCTION "public"."update_updated_at_column"();


--
-- TOC entry 3606 (class 2606 OID 25129)
-- Name: audit_logs audit_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."audit_logs"
    ADD CONSTRAINT "audit_logs_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id");


--
-- TOC entry 3608 (class 2606 OID 25809)
-- Name: conversation_members conversation_members_conversation_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversation_members"
    ADD CONSTRAINT "conversation_members_conversation_fk" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations"("id") ON DELETE CASCADE;


--
-- TOC entry 3609 (class 2606 OID 25814)
-- Name: conversation_members conversation_members_user_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversation_members"
    ADD CONSTRAINT "conversation_members_user_fk" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3607 (class 2606 OID 25774)
-- Name: conversations conversations_owner_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."conversations"
    ADD CONSTRAINT "conversations_owner_fk" FOREIGN KEY ("owner_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3612 (class 2606 OID 25850)
-- Name: message_history message_history_message_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_history"
    ADD CONSTRAINT "message_history_message_fk" FOREIGN KEY ("message_id") REFERENCES "public"."messages"("id") ON DELETE CASCADE;


--
-- TOC entry 3613 (class 2606 OID 25855)
-- Name: message_history message_history_user_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_history"
    ADD CONSTRAINT "message_history_user_fk" FOREIGN KEY ("edited_by") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3610 (class 2606 OID 25830)
-- Name: message_mentions message_mentions_message_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_mentions"
    ADD CONSTRAINT "message_mentions_message_fk" FOREIGN KEY ("message_id") REFERENCES "public"."messages"("id") ON DELETE CASCADE;


--
-- TOC entry 3611 (class 2606 OID 25835)
-- Name: message_mentions message_mentions_user_fk; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_mentions"
    ADD CONSTRAINT "message_mentions_user_fk" FOREIGN KEY ("mentioned_user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3598 (class 2606 OID 25017)
-- Name: message_reactions message_reactions_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_reactions"
    ADD CONSTRAINT "message_reactions_message_id_fkey" FOREIGN KEY ("message_id") REFERENCES "public"."messages"("id") ON DELETE CASCADE;


--
-- TOC entry 3599 (class 2606 OID 25022)
-- Name: message_reactions message_reactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."message_reactions"
    ADD CONSTRAINT "message_reactions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3590 (class 2606 OID 24974)
-- Name: messages messages_from_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_from_user_fkey" FOREIGN KEY ("author_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3591 (class 2606 OID 25893)
-- Name: messages messages_parent_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_parent_message_id_fkey" FOREIGN KEY ("parent_message_id") REFERENCES "public"."messages"("id") ON DELETE SET NULL;


--
-- TOC entry 3592 (class 2606 OID 25140)
-- Name: messages messages_reply_to_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_reply_to_id_fkey" FOREIGN KEY ("reply_to_id") REFERENCES "public"."messages"("id");


--
-- TOC entry 3593 (class 2606 OID 24979)
-- Name: messages messages_to_user_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."messages"
    ADD CONSTRAINT "messages_to_user_fkey" FOREIGN KEY ("to_user") REFERENCES "public"."users"("id") ON DELETE SET NULL;


--
-- TOC entry 3604 (class 2606 OID 25088)
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."notifications"
    ADD CONSTRAINT "notifications_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3594 (class 2606 OID 24853)
-- Name: product_documents product_documents_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."product_documents"
    ADD CONSTRAINT "product_documents_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "public"."products"("id") ON DELETE CASCADE;


--
-- TOC entry 3595 (class 2606 OID 24969)
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY "public"."refresh_tokens"
    ADD CONSTRAINT "refresh_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3602 (class 2606 OID 25064)
-- Name: room_members room_members_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."room_members"
    ADD CONSTRAINT "room_members_room_id_fkey" FOREIGN KEY ("room_id") REFERENCES "public"."rooms"("id") ON DELETE CASCADE;


--
-- TOC entry 3603 (class 2606 OID 25069)
-- Name: room_members room_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."room_members"
    ADD CONSTRAINT "room_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3596 (class 2606 OID 25000)
-- Name: sanctions sanctions_moderator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."sanctions"
    ADD CONSTRAINT "sanctions_moderator_id_fkey" FOREIGN KEY ("moderator_id") REFERENCES "public"."users"("id");


--
-- TOC entry 3597 (class 2606 OID 24995)
-- Name: sanctions sanctions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."sanctions"
    ADD CONSTRAINT "sanctions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3600 (class 2606 OID 25044)
-- Name: user_blocks user_blocks_blocked_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_blocks"
    ADD CONSTRAINT "user_blocks_blocked_id_fkey" FOREIGN KEY ("blocked_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3601 (class 2606 OID 25039)
-- Name: user_blocks user_blocks_blocker_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_blocks"
    ADD CONSTRAINT "user_blocks_blocker_id_fkey" FOREIGN KEY ("blocker_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3605 (class 2606 OID 25110)
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: veza
--

ALTER TABLE ONLY "public"."user_sessions"
    ADD CONSTRAINT "user_sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users"("id") ON DELETE CASCADE;


--
-- TOC entry 3825 (class 0 OID 0)
-- Dependencies: 232
-- Name: TABLE "offers"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE "public"."offers" TO "veza";


--
-- TOC entry 3827 (class 0 OID 0)
-- Dependencies: 231
-- Name: SEQUENCE "offers_id_seq"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE "public"."offers_id_seq" TO "veza";


--
-- TOC entry 3829 (class 0 OID 0)
-- Dependencies: 234
-- Name: TABLE "products"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE "public"."products" TO "veza";


--
-- TOC entry 3831 (class 0 OID 0)
-- Dependencies: 233
-- Name: SEQUENCE "products_id_seq1"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE "public"."products_id_seq1" TO "veza";


--
-- TOC entry 3832 (class 0 OID 0)
-- Dependencies: 242
-- Name: TABLE "refresh_tokens"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE "public"."refresh_tokens" TO "veza";


--
-- TOC entry 3834 (class 0 OID 0)
-- Dependencies: 241
-- Name: SEQUENCE "refresh_tokens_id_seq"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE "public"."refresh_tokens_id_seq" TO "veza";


--
-- TOC entry 3841 (class 0 OID 0)
-- Dependencies: 230
-- Name: TABLE "tags"; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE "public"."tags" TO "veza";


--
-- TOC entry 2208 (class 826 OID 24754)
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "veza";


--
-- TOC entry 2209 (class 826 OID 24755)
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "veza";


-- Completed on 2025-06-21 09:37:25 UTC

--
-- PostgreSQL database dump complete
--

