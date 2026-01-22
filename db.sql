--
-- PostgreSQL database dump
--

-- Dumped from database version 15.15
-- Dumped by pg_dump version 17.5 (Debian 17.5-1)

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
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: bookmark_list_code; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.bookmark_list_code AS ENUM (
    'reading',
    'planned',
    'dropped',
    'completed',
    'favorites'
);


ALTER TYPE public.bookmark_list_code OWNER TO novels;

--
-- Name: edit_field_type; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.edit_field_type AS ENUM (
    'title',
    'alt_titles',
    'description',
    'author',
    'cover_url',
    'release_year',
    'original_chapters_count',
    'genres',
    'tags',
    'translation_status'
);


ALTER TYPE public.edit_field_type OWNER TO novels;

--
-- Name: edit_request_status; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.edit_request_status AS ENUM (
    'pending',
    'approved',
    'rejected',
    'withdrawn'
);


ALTER TYPE public.edit_request_status OWNER TO novels;

--
-- Name: news_category; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.news_category AS ENUM (
    'announcement',
    'update',
    'event',
    'community',
    'translation'
);


ALTER TYPE public.news_category OWNER TO novels;

--
-- Name: proposal_status; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.proposal_status AS ENUM (
    'draft',
    'moderation',
    'voting',
    'accepted',
    'rejected',
    'translating'
);


ALTER TYPE public.proposal_status OWNER TO novels;

--
-- Name: subscription_status; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.subscription_status AS ENUM (
    'active',
    'canceled',
    'past_due',
    'expired'
);


ALTER TYPE public.subscription_status OWNER TO novels;

--
-- Name: ticket_type; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.ticket_type AS ENUM (
    'daily_vote',
    'novel_request',
    'translation_ticket'
);


ALTER TYPE public.ticket_type OWNER TO novels;

--
-- Name: translation_status; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.translation_status AS ENUM (
    'ongoing',
    'completed',
    'paused',
    'dropped'
);


ALTER TYPE public.translation_status OWNER TO novels;

--
-- Name: user_role; Type: TYPE; Schema: public; Owner: novels
--

CREATE TYPE public.user_role AS ENUM (
    'guest',
    'user',
    'premium',
    'moderator',
    'admin'
);


ALTER TYPE public.user_role OWNER TO novels;

--
-- Name: apply_edit_request(uuid, uuid); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.apply_edit_request(p_request_id uuid, p_moderator_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_request novel_edit_requests%ROWTYPE;
    v_change RECORD;
BEGIN
    -- Получаем запрос
    SELECT * INTO v_request FROM novel_edit_requests WHERE id = p_request_id AND status = 'pending';
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Применяем каждое изменение
    FOR v_change IN 
        SELECT * FROM novel_edit_request_changes WHERE request_id = p_request_id
    LOOP
        -- Записываем в историю
        INSERT INTO novel_edit_history (novel_id, request_id, user_id, field_type, lang, old_value, new_value)
        VALUES (v_request.novel_id, p_request_id, v_request.user_id, v_change.field_type, v_change.lang, v_change.old_value, v_change.new_value);
        
        -- Применяем изменение в зависимости от типа поля
        CASE v_change.field_type
            WHEN 'title' THEN
                IF v_change.lang IS NOT NULL THEN
                    UPDATE novel_localizations SET title = v_change.new_value, updated_at = NOW()
                    WHERE novel_id = v_request.novel_id AND lang = v_change.lang;
                END IF;
            WHEN 'description' THEN
                IF v_change.lang IS NOT NULL THEN
                    UPDATE novel_localizations SET description = v_change.new_value, updated_at = NOW()
                    WHERE novel_id = v_request.novel_id AND lang = v_change.lang;
                END IF;
            WHEN 'cover_url' THEN
                UPDATE novels SET cover_url = v_change.new_value, updated_at = NOW()
                WHERE id = v_request.novel_id;
            WHEN 'release_year' THEN
                UPDATE novels SET release_year = CAST(v_change.new_value AS INTEGER), updated_at = NOW()
                WHERE id = v_request.novel_id;
            WHEN 'original_chapters_count' THEN
                UPDATE novels SET original_chapters_count = CAST(v_change.new_value AS INTEGER), updated_at = NOW()
                WHERE id = v_request.novel_id;
            WHEN 'translation_status' THEN
                UPDATE novels SET translation_status = v_change.new_value::translation_status, updated_at = NOW()
                WHERE id = v_request.novel_id;
            ELSE
                -- Другие типы обрабатываем отдельно в приложении
                NULL;
        END CASE;
    END LOOP;
    
    -- Обновляем статус запроса
    UPDATE novel_edit_requests 
    SET status = 'approved', moderator_id = p_moderator_id, reviewed_at = NOW(), updated_at = NOW()
    WHERE id = p_request_id;
    
    RETURN TRUE;
END;
$$;


ALTER FUNCTION public.apply_edit_request(p_request_id uuid, p_moderator_id uuid) OWNER TO novels;

--
-- Name: get_setting(character varying); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.get_setting(p_key character varying) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_value JSONB;
BEGIN
    SELECT value INTO v_value FROM app_settings WHERE key = p_key;
    RETURN v_value;
END;
$$;


ALTER FUNCTION public.get_setting(p_key character varying) OWNER TO novels;

--
-- Name: log_admin_action(uuid, character varying, character varying, uuid, jsonb, character varying, text); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.log_admin_action(p_actor_user_id uuid, p_action character varying, p_entity_type character varying, p_entity_id uuid DEFAULT NULL::uuid, p_details jsonb DEFAULT NULL::jsonb, p_ip_address character varying DEFAULT NULL::character varying, p_user_agent text DEFAULT NULL::text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_log_id UUID;
BEGIN
    INSERT INTO admin_audit_log (
        actor_user_id,
        action,
        entity_type,
        entity_id,
        details,
        ip_address,
        user_agent
    ) VALUES (
        p_actor_user_id,
        p_action,
        p_entity_type,
        p_entity_id,
        p_details,
        p_ip_address,
        p_user_agent
    ) RETURNING id INTO v_log_id;
    
    RETURN v_log_id;
END;
$$;


ALTER FUNCTION public.log_admin_action(p_actor_user_id uuid, p_action character varying, p_entity_type character varying, p_entity_id uuid, p_details jsonb, p_ip_address character varying, p_user_agent text) OWNER TO novels;

--
-- Name: refresh_platform_stats(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.refresh_platform_stats() RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE platform_stats SET
        total_novels = (SELECT COUNT(*) FROM novels),
        total_chapters = (SELECT COUNT(*) FROM chapters),
        total_users = (SELECT COUNT(*) FROM users WHERE is_banned = false),
        total_comments = (SELECT COUNT(*) FROM comments WHERE is_deleted = false),
        total_collections = (SELECT COUNT(*) FROM collections WHERE is_public = true),
        updated_at = NOW()
    WHERE id = 1;
END;
$$;


ALTER FUNCTION public.refresh_platform_stats() OWNER TO novels;

--
-- Name: set_comment_hierarchy(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.set_comment_hierarchy() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        -- Корневой комментарий
        NEW.root_id = NEW.id;
        NEW.depth = 0;
    ELSE
        -- Дочерний комментарий - берем root_id и depth от родителя
        SELECT root_id, depth + 1 INTO NEW.root_id, NEW.depth
        FROM comments
        WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_comment_hierarchy() OWNER TO novels;

--
-- Name: update_chapter_word_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_chapter_word_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.word_count = array_length(regexp_split_to_array(NEW.content, '\s+'), 1);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_chapter_word_count() OWNER TO novels;

--
-- Name: update_collection_items_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_collection_items_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE collections SET items_count = items_count + 1, updated_at = NOW()
        WHERE id = NEW.collection_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE collections SET items_count = items_count - 1, updated_at = NOW()
        WHERE id = OLD.collection_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_collection_items_count() OWNER TO novels;

--
-- Name: update_collection_votes_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_collection_votes_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE collections SET votes_count = votes_count + NEW.value, updated_at = NOW()
        WHERE id = NEW.collection_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE collections SET votes_count = votes_count - OLD.value, updated_at = NOW()
        WHERE id = OLD.collection_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_collection_votes_count() OWNER TO novels;

--
-- Name: update_comment_replies_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_comment_replies_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.parent_id IS NOT NULL AND NOT NEW.is_deleted THEN
        UPDATE comments SET replies_count = replies_count + 1 WHERE id = NEW.parent_id;
    ELSIF TG_OP = 'DELETE' AND OLD.parent_id IS NOT NULL THEN
        UPDATE comments SET replies_count = GREATEST(0, replies_count - 1) WHERE id = OLD.parent_id;
    ELSIF TG_OP = 'UPDATE' AND NEW.parent_id IS NOT NULL THEN
        -- Если комментарий был помечен как удаленный
        IF OLD.is_deleted = FALSE AND NEW.is_deleted = TRUE THEN
            UPDATE comments SET replies_count = GREATEST(0, replies_count - 1) WHERE id = NEW.parent_id;
        -- Если комментарий был восстановлен
        ELSIF OLD.is_deleted = TRUE AND NEW.is_deleted = FALSE THEN
            UPDATE comments SET replies_count = replies_count + 1 WHERE id = NEW.parent_id;
        END IF;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_comment_replies_count() OWNER TO novels;

--
-- Name: update_comment_vote_counts(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_comment_vote_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.value = 1 THEN
            UPDATE comments SET likes_count = likes_count + 1 WHERE id = NEW.comment_id;
        ELSE
            UPDATE comments SET dislikes_count = dislikes_count + 1 WHERE id = NEW.comment_id;
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.value = 1 THEN
            UPDATE comments SET likes_count = GREATEST(0, likes_count - 1) WHERE id = OLD.comment_id;
        ELSE
            UPDATE comments SET dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = OLD.comment_id;
        END IF;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.value != NEW.value THEN
            IF NEW.value = 1 THEN
                UPDATE comments SET likes_count = likes_count + 1, dislikes_count = GREATEST(0, dislikes_count - 1) WHERE id = NEW.comment_id;
            ELSE
                UPDATE comments SET likes_count = GREATEST(0, likes_count - 1), dislikes_count = dislikes_count + 1 WHERE id = NEW.comment_id;
            END IF;
        END IF;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_comment_vote_counts() OWNER TO novels;

--
-- Name: update_news_comments_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_news_comments_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.target_type = 'news' THEN
        UPDATE news_posts SET comments_count = comments_count + 1
        WHERE id = NEW.target_id;
    ELSIF TG_OP = 'DELETE' AND OLD.target_type = 'news' THEN
        UPDATE news_posts SET comments_count = comments_count - 1
        WHERE id = OLD.target_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_news_comments_count() OWNER TO novels;

--
-- Name: update_novel_bookmarks_count(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_novel_bookmarks_count() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE novels SET bookmarks_count = bookmarks_count + 1 WHERE id = NEW.novel_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE novels SET bookmarks_count = bookmarks_count - 1 WHERE id = OLD.novel_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_novel_bookmarks_count() OWNER TO novels;

--
-- Name: update_novel_rating(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_novel_rating() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE novels SET 
            rating_sum = rating_sum + NEW.value,
            rating_count = rating_count + 1
        WHERE id = NEW.novel_id;
    ELSIF TG_OP = 'UPDATE' THEN
        UPDATE novels SET 
            rating_sum = rating_sum - OLD.value + NEW.value
        WHERE id = NEW.novel_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE novels SET 
            rating_sum = rating_sum - OLD.value,
            rating_count = rating_count - 1
        WHERE id = OLD.novel_id;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_novel_rating() OWNER TO novels;

--
-- Name: update_novel_search_vector(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_novel_search_vector() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.search_vector = 
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(array_to_string(NEW.alt_titles, ' '), '')), 'A');
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_novel_search_vector() OWNER TO novels;

--
-- Name: update_proposal_stats(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_proposal_stats() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.ticket_type = 'daily_vote' THEN
            UPDATE novel_proposals
            SET vote_score = vote_score + NEW.amount,
                votes_count = votes_count + 1,
                updated_at = NOW()
            WHERE id = NEW.proposal_id;
        ELSIF NEW.ticket_type = 'translation_ticket' THEN
            UPDATE novel_proposals
            SET translation_tickets_invested = translation_tickets_invested + NEW.amount,
                updated_at = NOW()
            WHERE id = NEW.proposal_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF OLD.ticket_type = 'daily_vote' THEN
            UPDATE novel_proposals
            SET vote_score = vote_score - OLD.amount,
                votes_count = votes_count - 1,
                updated_at = NOW()
            WHERE id = OLD.proposal_id;
        ELSIF OLD.ticket_type = 'translation_ticket' THEN
            UPDATE novel_proposals
            SET translation_tickets_invested = translation_tickets_invested - OLD.amount,
                updated_at = NOW()
            WHERE id = OLD.proposal_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_proposal_stats() OWNER TO novels;

--
-- Name: update_proposal_vote_score(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_proposal_vote_score() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE novel_proposals
        SET vote_score = vote_score + NEW.amount,
            votes_count = votes_count + 1,
            updated_at = NOW()
        WHERE id = NEW.proposal_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE novel_proposals
        SET vote_score = vote_score - OLD.amount,
            votes_count = votes_count - 1,
            updated_at = NOW()
        WHERE id = OLD.proposal_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$;


ALTER FUNCTION public.update_proposal_vote_score() OWNER TO novels;

--
-- Name: update_setting(character varying, jsonb, uuid); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_setting(p_key character varying, p_value jsonb, p_updated_by uuid DEFAULT NULL::uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE app_settings 
    SET value = p_value, updated_by = p_updated_by, updated_at = NOW()
    WHERE key = p_key;
    
    RETURN FOUND;
END;
$$;


ALTER FUNCTION public.update_setting(p_key character varying, p_value jsonb, p_updated_by uuid) OWNER TO novels;

--
-- Name: update_subscription_timestamp(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_subscription_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_subscription_timestamp() OWNER TO novels;

--
-- Name: update_ticket_balance(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_ticket_balance() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO ticket_balances (user_id, type, balance, updated_at)
    VALUES (NEW.user_id, NEW.type, GREATEST(0, NEW.delta), NOW())
    ON CONFLICT (user_id, type)
    DO UPDATE SET
        balance = GREATEST(0, ticket_balances.balance + NEW.delta),
        updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_ticket_balance() OWNER TO novels;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: novels
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO novels;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: achievements; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.achievements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    icon_key character varying(100),
    condition jsonb,
    xp_reward bigint DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.achievements OWNER TO novels;

--
-- Name: subscription_plans; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.subscription_plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(50) NOT NULL,
    title character varying(100) NOT NULL,
    description text,
    price integer NOT NULL,
    currency character varying(3) DEFAULT 'RUB'::character varying NOT NULL,
    period character varying(20) DEFAULT 'monthly'::character varying NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    features jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.subscription_plans OWNER TO novels;

--
-- Name: TABLE subscription_plans; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.subscription_plans IS 'Available subscription plans with features';


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    status public.subscription_status DEFAULT 'active'::public.subscription_status NOT NULL,
    starts_at timestamp with time zone DEFAULT now() NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    external_id character varying(255),
    auto_renew boolean DEFAULT true NOT NULL,
    canceled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.subscriptions OWNER TO novels;

--
-- Name: TABLE subscriptions; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.subscriptions IS 'User subscriptions';


--
-- Name: active_subscriptions; Type: VIEW; Schema: public; Owner: novels
--

CREATE VIEW public.active_subscriptions AS
 SELECT s.id,
    s.user_id,
    s.plan_id,
    s.status,
    s.starts_at,
    s.ends_at,
    s.external_id,
    s.auto_renew,
    s.canceled_at,
    s.created_at,
    s.updated_at,
    sp.code AS plan_code,
    sp.title AS plan_title,
    sp.features
   FROM (public.subscriptions s
     JOIN public.subscription_plans sp ON ((s.plan_id = sp.id)))
  WHERE ((s.status = 'active'::public.subscription_status) AND (s.ends_at > now()));


ALTER VIEW public.active_subscriptions OWNER TO novels;

--
-- Name: admin_audit_log; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.admin_audit_log (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    actor_user_id uuid NOT NULL,
    action character varying(100) NOT NULL,
    entity_type character varying(50) NOT NULL,
    entity_id uuid,
    details jsonb,
    ip_address character varying(45),
    user_agent text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.admin_audit_log OWNER TO novels;

--
-- Name: app_settings; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.app_settings (
    key character varying(100) NOT NULL,
    value jsonb NOT NULL,
    description text,
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.app_settings OWNER TO novels;

--
-- Name: author_localizations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.author_localizations (
    author_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    name character varying(255) NOT NULL,
    bio text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.author_localizations OWNER TO novels;

--
-- Name: authors; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.authors (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    slug character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.authors OWNER TO novels;

--
-- Name: bookmark_lists; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.bookmark_lists (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    code public.bookmark_list_code NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.bookmark_lists OWNER TO novels;

--
-- Name: bookmarks; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.bookmarks (
    user_id uuid NOT NULL,
    novel_id uuid NOT NULL,
    list_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    id uuid DEFAULT gen_random_uuid()
);


ALTER TABLE public.bookmarks OWNER TO novels;

--
-- Name: chapter_contents; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.chapter_contents (
    chapter_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    content text NOT NULL,
    word_count integer DEFAULT 0 NOT NULL,
    source character varying(50) DEFAULT 'manual'::character varying NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.chapter_contents OWNER TO novels;

--
-- Name: chapters; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.chapters (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    novel_id uuid NOT NULL,
    number numeric(10,2) NOT NULL,
    slug character varying(255),
    title character varying(500),
    views integer DEFAULT 0 NOT NULL,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.chapters OWNER TO novels;

--
-- Name: collection_items; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.collection_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    collection_id uuid NOT NULL,
    novel_id uuid NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    note text,
    added_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.collection_items OWNER TO novels;

--
-- Name: collection_votes; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.collection_votes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    collection_id uuid NOT NULL,
    user_id uuid NOT NULL,
    value integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.collection_votes OWNER TO novels;

--
-- Name: collections; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.collections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    slug character varying(255) NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    cover_url text,
    is_public boolean DEFAULT true,
    is_featured boolean DEFAULT false,
    views_count integer DEFAULT 0,
    votes_count integer DEFAULT 0,
    items_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.collections OWNER TO novels;

--
-- Name: comment_reports; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.comment_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    comment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    reason text NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT comment_reports_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'resolved'::character varying, 'dismissed'::character varying])::text[])))
);


ALTER TABLE public.comment_reports OWNER TO novels;

--
-- Name: comment_votes; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.comment_votes (
    comment_id uuid NOT NULL,
    user_id uuid NOT NULL,
    value integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT comment_votes_value_check CHECK ((value = ANY (ARRAY['-1'::integer, 1])))
);


ALTER TABLE public.comment_votes OWNER TO novels;

--
-- Name: comments; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.comments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    novel_id uuid,
    chapter_id uuid,
    user_id uuid NOT NULL,
    parent_id uuid,
    content text NOT NULL,
    is_spoiler boolean DEFAULT false NOT NULL,
    likes_count integer DEFAULT 0 NOT NULL,
    dislikes_count integer DEFAULT 0 NOT NULL,
    replies_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    target_type character varying(20) NOT NULL,
    target_id uuid NOT NULL,
    body text NOT NULL,
    is_deleted boolean DEFAULT false NOT NULL,
    root_id uuid,
    depth integer DEFAULT 0 NOT NULL,
    anchor text
);


ALTER TABLE public.comments OWNER TO novels;

--
-- Name: COLUMN comments.novel_id; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON COLUMN public.comments.novel_id IS 'DEPRECATED: Use target_type=novel and target_id instead. Will be removed in future migration.';


--
-- Name: COLUMN comments.chapter_id; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON COLUMN public.comments.chapter_id IS 'DEPRECATED: Use target_type=chapter and target_id instead. Will be removed in future migration.';


--
-- Name: COLUMN comments.content; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON COLUMN public.comments.content IS 'DEPRECATED: Use body instead. Will be removed in future migration.';


--
-- Name: COLUMN comments.deleted_at; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON COLUMN public.comments.deleted_at IS 'DEPRECATED: Use is_deleted instead. Will be removed in future migration.';


--
-- Name: daily_vote_grants; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.daily_vote_grants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    grant_date date NOT NULL,
    users_processed integer DEFAULT 0 NOT NULL,
    total_votes_granted integer DEFAULT 0 NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    status character varying(20) DEFAULT 'running'::character varying NOT NULL,
    error_message text
);


ALTER TABLE public.daily_vote_grants OWNER TO novels;

--
-- Name: TABLE daily_vote_grants; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.daily_vote_grants IS 'Logs of daily vote grant executions';


--
-- Name: genre_localizations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.genre_localizations (
    genre_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.genre_localizations OWNER TO novels;

--
-- Name: genres; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.genres (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    slug character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.genres OWNER TO novels;

--
-- Name: leaderboard_cache; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.leaderboard_cache (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    period character varying(10) NOT NULL,
    user_id uuid NOT NULL,
    tickets_spent integer DEFAULT 0 NOT NULL,
    rank integer DEFAULT 0 NOT NULL,
    calculated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.leaderboard_cache OWNER TO novels;

--
-- Name: TABLE leaderboard_cache; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.leaderboard_cache IS 'Cached leaderboard data for performance';


--
-- Name: news_localizations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.news_localizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    news_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    title character varying(500) NOT NULL,
    summary text,
    content text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.news_localizations OWNER TO novels;

--
-- Name: news_posts; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.news_posts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    slug character varying(255) NOT NULL,
    title character varying(500) NOT NULL,
    summary text,
    content text NOT NULL,
    cover_url text,
    category public.news_category DEFAULT 'announcement'::public.news_category,
    author_id uuid NOT NULL,
    is_published boolean DEFAULT false,
    is_pinned boolean DEFAULT false,
    views_count integer DEFAULT 0,
    comments_count integer DEFAULT 0,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.news_posts OWNER TO novels;

--
-- Name: novel_authors; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_authors (
    novel_id uuid NOT NULL,
    author_id uuid NOT NULL,
    is_primary boolean DEFAULT true NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.novel_authors OWNER TO novels;

--
-- Name: novel_edit_history; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_edit_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    novel_id uuid NOT NULL,
    request_id uuid,
    user_id uuid NOT NULL,
    field_type public.edit_field_type NOT NULL,
    lang character varying(10),
    old_value text,
    new_value text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.novel_edit_history OWNER TO novels;

--
-- Name: novel_edit_request_changes; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_edit_request_changes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_id uuid NOT NULL,
    field_type public.edit_field_type NOT NULL,
    lang character varying(10),
    old_value text,
    new_value text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.novel_edit_request_changes OWNER TO novels;

--
-- Name: novel_edit_requests; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_edit_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    novel_id uuid NOT NULL,
    user_id uuid NOT NULL,
    status public.edit_request_status DEFAULT 'pending'::public.edit_request_status,
    edit_reason text,
    moderator_id uuid,
    moderator_comment text,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.novel_edit_requests OWNER TO novels;

--
-- Name: novel_genres; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_genres (
    novel_id uuid NOT NULL,
    genre_id uuid NOT NULL
);


ALTER TABLE public.novel_genres OWNER TO novels;

--
-- Name: novel_localizations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_localizations (
    novel_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    title character varying(500) NOT NULL,
    description text,
    alt_titles text[],
    search_vector tsvector,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.novel_localizations OWNER TO novels;

--
-- Name: novel_proposals; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_proposals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    original_link text NOT NULL,
    status public.proposal_status DEFAULT 'draft'::public.proposal_status NOT NULL,
    title character varying(500) NOT NULL,
    alt_titles text[] DEFAULT ARRAY[]::text[],
    author character varying(255),
    description text,
    cover_url text,
    genres text[] DEFAULT ARRAY[]::text[],
    tags text[] DEFAULT ARRAY[]::text[],
    vote_score integer DEFAULT 0 NOT NULL,
    votes_count integer DEFAULT 0 NOT NULL,
    moderator_id uuid,
    reject_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    translation_tickets_invested integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.novel_proposals OWNER TO novels;

--
-- Name: TABLE novel_proposals; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.novel_proposals IS 'User-submitted proposals for novels to translate';


--
-- Name: novel_ratings; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_ratings (
    novel_id uuid NOT NULL,
    user_id uuid NOT NULL,
    value integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT novel_ratings_value_check CHECK (((value >= 1) AND (value <= 10)))
);


ALTER TABLE public.novel_ratings OWNER TO novels;

--
-- Name: novel_tags; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_tags (
    novel_id uuid NOT NULL,
    tag_id uuid NOT NULL
);


ALTER TABLE public.novel_tags OWNER TO novels;

--
-- Name: novel_views_daily; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novel_views_daily (
    novel_id uuid NOT NULL,
    date date NOT NULL,
    views integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.novel_views_daily OWNER TO novels;

--
-- Name: novels; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.novels (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    slug character varying(255) NOT NULL,
    cover_image_key character varying(255),
    translation_status public.translation_status DEFAULT 'ongoing'::public.translation_status NOT NULL,
    original_chapters_count integer DEFAULT 0,
    release_year integer,
    author character varying(255),
    views_total bigint DEFAULT 0 NOT NULL,
    views_daily integer DEFAULT 0 NOT NULL,
    rating_sum integer DEFAULT 0 NOT NULL,
    rating_count integer DEFAULT 0 NOT NULL,
    bookmarks_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.novels OWNER TO novels;

--
-- Name: COLUMN novels.author; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON COLUMN public.novels.author IS 'DEPRECATED: Use novel_authors table instead. Will be removed in future migration.';


--
-- Name: platform_stats; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.platform_stats (
    id integer DEFAULT 1 NOT NULL,
    total_novels integer DEFAULT 0,
    total_chapters integer DEFAULT 0,
    total_users integer DEFAULT 0,
    total_comments integer DEFAULT 0,
    total_collections integer DEFAULT 0,
    total_votes_cast integer DEFAULT 0,
    total_tickets_spent bigint DEFAULT 0,
    proposals_translated integer DEFAULT 0,
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT single_row CHECK ((id = 1))
);


ALTER TABLE public.platform_stats OWNER TO novels;

--
-- Name: reading_progress; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.reading_progress (
    user_id uuid NOT NULL,
    novel_id uuid NOT NULL,
    chapter_id uuid NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.reading_progress OWNER TO novels;

--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.refresh_tokens (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone
);


ALTER TABLE public.refresh_tokens OWNER TO novels;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL,
    applied_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.schema_migrations OWNER TO novels;

--
-- Name: subscription_grants; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.subscription_grants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    subscription_id uuid NOT NULL,
    user_id uuid NOT NULL,
    type public.ticket_type NOT NULL,
    amount integer NOT NULL,
    for_month character varying(7) NOT NULL,
    granted_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.subscription_grants OWNER TO novels;

--
-- Name: TABLE subscription_grants; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.subscription_grants IS 'Monthly ticket grants from subscriptions';


--
-- Name: tag_localizations; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.tag_localizations (
    tag_id uuid NOT NULL,
    lang character varying(10) NOT NULL,
    name character varying(100) NOT NULL
);


ALTER TABLE public.tag_localizations OWNER TO novels;

--
-- Name: tags; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.tags (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    slug character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tags OWNER TO novels;

--
-- Name: ticket_balances; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.ticket_balances (
    user_id uuid NOT NULL,
    type public.ticket_type NOT NULL,
    balance integer DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ticket_balances_balance_check CHECK ((balance >= 0))
);


ALTER TABLE public.ticket_balances OWNER TO novels;

--
-- Name: TABLE ticket_balances; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.ticket_balances IS 'Current ticket balances for each user and ticket type';


--
-- Name: ticket_transactions; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.ticket_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type public.ticket_type NOT NULL,
    delta integer NOT NULL,
    reason character varying(100) NOT NULL,
    ref_type character varying(50),
    ref_id uuid,
    idempotency_key character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.ticket_transactions OWNER TO novels;

--
-- Name: TABLE ticket_transactions; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.ticket_transactions IS 'All ticket transactions (credits and debits)';


--
-- Name: user_achievements; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.user_achievements (
    user_id uuid NOT NULL,
    achievement_id uuid NOT NULL,
    unlocked_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_achievements OWNER TO novels;

--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.user_profiles (
    user_id uuid NOT NULL,
    display_name character varying(100) NOT NULL,
    avatar_key character varying(255),
    bio text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_profiles OWNER TO novels;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.user_roles (
    user_id uuid NOT NULL,
    role public.user_role NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_roles OWNER TO novels;

--
-- Name: users; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    is_banned boolean DEFAULT false NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.users OWNER TO novels;

--
-- Name: user_wallets; Type: VIEW; Schema: public; Owner: novels
--

CREATE VIEW public.user_wallets AS
 SELECT u.id AS user_id,
    COALESCE(dv.balance, 0) AS daily_votes,
    COALESCE(nr.balance, 0) AS novel_requests,
    COALESCE(tt.balance, 0) AS translation_tickets
   FROM (((public.users u
     LEFT JOIN public.ticket_balances dv ON (((u.id = dv.user_id) AND (dv.type = 'daily_vote'::public.ticket_type))))
     LEFT JOIN public.ticket_balances nr ON (((u.id = nr.user_id) AND (nr.type = 'novel_request'::public.ticket_type))))
     LEFT JOIN public.ticket_balances tt ON (((u.id = tt.user_id) AND (tt.type = 'translation_ticket'::public.ticket_type))));


ALTER VIEW public.user_wallets OWNER TO novels;

--
-- Name: user_xp; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.user_xp (
    user_id uuid NOT NULL,
    xp_total bigint DEFAULT 0 NOT NULL,
    level integer DEFAULT 1 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.user_xp OWNER TO novels;

--
-- Name: votes; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.votes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    poll_id uuid,
    user_id uuid NOT NULL,
    proposal_id uuid NOT NULL,
    ticket_type public.ticket_type NOT NULL,
    amount integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT votes_amount_check CHECK ((amount > 0))
);


ALTER TABLE public.votes OWNER TO novels;

--
-- Name: TABLE votes; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.votes IS 'Individual votes cast by users';


--
-- Name: voting_leaderboard; Type: VIEW; Schema: public; Owner: novels
--

CREATE VIEW public.voting_leaderboard AS
 SELECT np.id,
    np.user_id,
    np.original_link,
    np.status,
    np.title,
    np.alt_titles,
    np.author,
    np.description,
    np.cover_url,
    np.genres,
    np.tags,
    np.vote_score,
    np.votes_count,
    np.moderator_id,
    np.reject_reason,
    np.created_at,
    np.updated_at,
    u.email AS user_email,
    up.display_name AS user_display_name,
    up.avatar_key AS user_avatar,
    ux.level AS user_level
   FROM (((public.novel_proposals np
     JOIN public.users u ON ((np.user_id = u.id)))
     LEFT JOIN public.user_profiles up ON ((np.user_id = up.user_id)))
     LEFT JOIN public.user_xp ux ON ((np.user_id = ux.user_id)))
  WHERE (np.status = 'voting'::public.proposal_status)
  ORDER BY np.vote_score DESC, np.created_at;


ALTER VIEW public.voting_leaderboard OWNER TO novels;

--
-- Name: voting_polls; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.voting_polls (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying NOT NULL,
    starts_at timestamp with time zone DEFAULT now() NOT NULL,
    ends_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.voting_polls OWNER TO novels;

--
-- Name: TABLE voting_polls; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.voting_polls IS 'Voting periods for novel selection';


--
-- Name: weekly_ticket_grants; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.weekly_ticket_grants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    grant_date date NOT NULL,
    users_processed integer DEFAULT 0 NOT NULL,
    novel_requests_granted integer DEFAULT 0 NOT NULL,
    translation_tickets_granted integer DEFAULT 0 NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    status character varying(20) DEFAULT 'running'::character varying NOT NULL,
    error_message text
);


ALTER TABLE public.weekly_ticket_grants OWNER TO novels;

--
-- Name: TABLE weekly_ticket_grants; Type: COMMENT; Schema: public; Owner: novels
--

COMMENT ON TABLE public.weekly_ticket_grants IS 'Logs of weekly ticket grant executions (Wed 00:00 UTC)';


--
-- Name: xp_events; Type: TABLE; Schema: public; Owner: novels
--

CREATE TABLE public.xp_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    delta bigint NOT NULL,
    ref_type character varying(50),
    ref_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.xp_events OWNER TO novels;

--
-- Data for Name: achievements; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.achievements (id, code, title, description, icon_key, condition, xp_reward, created_at) FROM stdin;
8a2a9f93-0d1f-44be-b187-103cc44896cd	first_chapter	Первая глава	Прочитать первую главу	book-open	\N	50	2026-01-11 16:04:08.191329+00
a4044c0e-63e6-46a6-bd64-0c3117a11814	bookworm_10	Книжный червь	Прочитать 10 глав	book	\N	100	2026-01-11 16:04:08.191329+00
1c66ef57-9d6f-43f5-9ed4-bf7d5d8e7888	bookworm_100	Заядлый читатель	Прочитать 100 глав	books	\N	500	2026-01-11 16:04:08.191329+00
9fbcec1a-5a1b-48a2-a63c-0b554e38a238	bookworm_1000	Мастер чтения	Прочитать 1000 глав	crown	\N	2000	2026-01-11 16:04:08.191329+00
c446faa5-3a12-409a-b8ff-9a24f57ad167	first_comment	Первый комментарий	Оставить первый комментарий	message	\N	25	2026-01-11 16:04:08.191329+00
2e3ae73c-9983-4f72-8dbe-c0f68fb3e000	commentator_10	Комментатор	Оставить 10 комментариев	messages	\N	100	2026-01-11 16:04:08.191329+00
40c89a1a-d818-4d22-8528-ef0b1278f4da	commentator_100	Активный участник	Оставить 100 комментариев	chat	\N	500	2026-01-11 16:04:08.191329+00
45613ced-8b0d-44a4-ad2d-cd59e66a94e7	first_bookmark	Первая закладка	Добавить первую книгу в закладки	bookmark	\N	15	2026-01-11 16:04:08.191329+00
b629f6b2-00e2-40f1-a921-f30cf82f8523	collector_10	Коллекционер	Добавить 10 книг в закладки	bookmarks	\N	75	2026-01-11 16:04:08.191329+00
de19cc9b-1861-4d56-8b1f-634eff8aab26	collector_100	Библиотекарь	Добавить 100 книг в закладки	library	\N	300	2026-01-11 16:04:08.191329+00
\.


--
-- Data for Name: admin_audit_log; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.admin_audit_log (id, actor_user_id, action, entity_type, entity_id, details, ip_address, user_agent, created_at) FROM stdin;
\.


--
-- Data for Name: app_settings; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.app_settings (key, value, description, updated_by, updated_at) FROM stdin;
site_name	{"en": "Novels", "ru": "Новеллы"}	Название сайта	\N	2026-01-16 22:42:07.900141+00
site_description	{"en": "Platform for reading novels", "ru": "Платформа для чтения новелл"}	Описание сайта	\N	2026-01-16 22:42:07.900141+00
registration_enabled	true	Разрешена ли регистрация новых пользователей	\N	2026-01-16 22:42:07.900141+00
comments_enabled	true	Включены ли комментарии на сайте	\N	2026-01-16 22:42:07.900141+00
maintenance_mode	false	Режим технического обслуживания	\N	2026-01-16 22:42:07.900141+00
max_upload_size_mb	10	Максимальный размер загружаемого файла в МБ	\N	2026-01-16 22:42:07.900141+00
chapters_per_day_limit	0	Лимит глав в день для обычных пользователей (0 = без лимита)	\N	2026-01-16 22:42:07.900141+00
xp_multiplier	1.0	Множитель для начисления XP	\N	2026-01-16 22:42:07.900141+00
featured_novels_count	6	Количество рекомендуемых новелл на главной	\N	2026-01-16 22:42:07.900141+00
trending_days_window	7	Окно в днях для расчета трендовых новелл	\N	2026-01-16 22:42:07.900141+00
comments_per_page	20	Количество комментариев на странице	\N	2026-01-16 22:42:07.900141+00
novels_per_page	24	Количество новелл на странице каталога	\N	2026-01-16 22:42:07.900141+00
min_comment_length	1	Минимальная длина комментария	\N	2026-01-16 22:42:07.900141+00
max_comment_length	10000	Максимальная длина комментария	\N	2026-01-16 22:42:07.900141+00
email_verification_required	false	Требуется ли подтверждение email при регистрации	\N	2026-01-16 22:42:07.900141+00
analytics_enabled	false	Включена ли аналитика	\N	2026-01-16 22:42:07.900141+00
cdn_url	null	URL CDN для статических файлов	\N	2026-01-16 22:42:07.900141+00
backup_retention_days	30	Количество дней хранения резервных копий	\N	2026-01-16 22:42:07.900141+00
\.


--
-- Data for Name: author_localizations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.author_localizations (author_id, lang, name, bio, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: authors; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.authors (id, slug, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: bookmark_lists; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.bookmark_lists (id, user_id, code, sort_order, created_at) FROM stdin;
cb50153c-958b-410c-80b2-1ddfbc1bce77	ce76495c-3f80-4c7a-be56-3d5994b49b03	reading	0	2026-01-11 21:31:49.403422+00
9d3f7d2e-7b5e-457c-bf85-d0419ef57edc	ce76495c-3f80-4c7a-be56-3d5994b49b03	planned	1	2026-01-11 21:31:49.403422+00
d3c0808c-1280-4626-b2c9-ca51d49e38ec	ce76495c-3f80-4c7a-be56-3d5994b49b03	dropped	2	2026-01-11 21:31:49.403422+00
31b2dc50-59a7-48c2-b30a-5cf442e6ea38	ce76495c-3f80-4c7a-be56-3d5994b49b03	completed	3	2026-01-11 21:31:49.403422+00
1b684f03-6e0f-4516-b4e9-ceb296733621	ce76495c-3f80-4c7a-be56-3d5994b49b03	favorites	4	2026-01-11 21:31:49.403422+00
\.


--
-- Data for Name: bookmarks; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.bookmarks (user_id, novel_id, list_id, created_at, updated_at, id) FROM stdin;
\.


--
-- Data for Name: chapter_contents; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.chapter_contents (chapter_id, lang, content, word_count, source, updated_at) FROM stdin;
7427c8b3-d65d-407a-b8a2-b2400a93183a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
87a006b4-edb9-425b-9d9c-57de8a18ab5f	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
9bb21178-f58d-4821-bf1b-10a23e175949	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
29bd06c1-7c59-44f6-8293-c06ec06f480f	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
c27f3c04-4436-4fc2-b0c8-6cf169a2a87d	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
93941aa7-b3cf-4b0f-a0ed-0f76bbba2f12	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
31a2686c-6723-4ee6-8bde-834048f20990	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
b535bbee-5632-4c24-bd14-d84c52bdb29e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
ca7db5ff-526e-4106-a1d6-16215c738054	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
1b1426b2-c8ab-455c-a8b9-8aa7f70cda06	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
8b11fc6a-25db-4d9a-9c80-8060828f983e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
8bda608d-9b30-4a7f-9adb-3cf4cbd84b13	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
ac5da8cd-bf46-49fa-9ee5-9e4b2efacca5	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
1bffb6b2-26d1-42e6-9a22-e7870091bc94	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
5fa92323-a25c-4413-9b73-07c83d732952	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3c1266e1-1f79-4510-994e-43a6a77a3bf7	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
5e940078-ea05-4602-a57d-5aeda861d035	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
964cbcd7-be1c-4efb-a290-c7ca0f0859ed	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
c1a36b57-6ed2-4cf1-8760-bb8d27dbadb1	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
9516f965-36a6-4cb4-b665-5f2c984d721e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
857430f8-b9cd-4105-a4da-e749e66ed09a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3687251c-efd4-40a5-9b3e-4cf98ad2bf4c	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
2ffcc8c5-14b4-46a8-9cd9-676b7082ce06	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
25088950-e2f8-4028-bd98-0e5f116cbb3d	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
31c50974-7e2d-4ba9-9dd5-909647491016	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3eae23d3-eb46-46f8-8ed5-c6e3d6b2c00a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
cc28095d-723e-48e7-bf84-47661443dc59	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
9948543a-3338-4d0e-8ec7-34a0713b39f1	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
dc5573a7-accb-4d69-8df6-b848f38b487c	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
e15fcca1-ae5c-45c3-ab52-1699ac4e3b3c	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
cf9b1825-fdd5-4a21-80ec-d1eea8547782	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
c0c6aff3-9a5e-4826-ae59-3afc3eceb5ed	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
f62a7132-a58e-4ebc-9b11-7281803c2562	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
dc1b189d-b2ae-4af9-bb86-3160373af705	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
2a05c44c-7664-43bd-8836-d199ec58379a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
b06ff097-18f9-45d3-9fbd-9392d3bbf8f7	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
13cac42f-efca-41e2-be2c-c8714e7f748b	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
eba42a24-906f-46c9-918e-a9a07203b997	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
e9b0c65d-2d27-4c9e-a4a1-a248d2e1463e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
2c546d05-11ef-4a22-bef1-43c87139bc26	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3ce08b80-2557-48ee-ad09-263e16cfccad	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
6bbb1826-6488-4ceb-b91b-c21e20f21b4a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
0227ad5c-a8f5-4b29-a982-df3188f9b4eb	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
077644e0-20a2-4e97-9ac9-71708670dd4c	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
b1721dd0-8679-40a4-85a0-de4690136200	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
050f1e13-9af7-447c-9f4a-a9b6859c97a0	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
15cead81-2436-4638-80dd-7665bef242ca	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
64ed73a6-ef3f-4c2e-8508-04c0762a6114	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
7a5b9278-a963-472c-936b-e6b9ad9a3d92	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
79ce09e5-164f-4b3a-b702-d8258e74a3b9	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
cf82adda-1a6c-407a-8b17-0613077e6f26	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
a80b5117-751d-41eb-894e-a6fc941c2a9a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
43afbd68-5759-4a5a-84e6-acb0165f2dc6	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3059c97e-ed04-4535-8dc9-e2e2081273a7	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
262602d2-e610-47e0-97b9-651b61a081c6	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
8f88a42e-930f-43cf-8826-2682f4626393	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
4bb71bb7-9aeb-4d0e-ab14-7a9df1f72679	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
9aff976f-6566-4dc7-b929-ee7ea2de891a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
bcdced7c-c60b-4ac5-90e6-142ea8069c60	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
c62858e8-3394-4292-a6c1-508842c5ca45	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
772b6122-3540-4e37-b7dd-35fc1ab177c0	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
5e7b67d5-98ce-4d13-8bca-c20012e692c9	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
64856e08-d899-46e8-901e-c76af22d9e65	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
268692ad-f1e9-4e54-8963-36ef12b5262f	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
7380d62c-1fa0-4516-be36-e794f7f8f73c	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
12f52664-5a3a-4465-8a45-1650d71f69eb	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
f8f19ddf-a8fa-43fe-bb8b-00b5adfc4922	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
89f6d36a-71c9-4e10-b5c8-6b323712bf81	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
fd24295e-d635-49ad-a28a-928ea5cbb5ee	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
61c5a2e3-94f8-4460-a80d-cc29aa1197f5	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
6e46fd81-ecfe-4759-9757-48028041d27b	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
e7650081-c532-47eb-a9f5-b0f31bb50319	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
c55e7ece-7bac-428f-9b35-036684eed7ae	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
35e7f90c-a17a-4280-b803-1b56d3fbd33e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
beb21418-fa55-4af2-b2e2-8c83f4bca26e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
eede3327-1a63-4d17-957c-2c30423415d4	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
e40edea1-5280-48eb-ac4e-ab626034a442	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
aa5cd215-c1de-4b9c-aaf5-fdcd71344e10	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
41a11632-cf61-462b-a7e8-f9dcfa6334e3	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
63872525-4b07-4f5a-a2f3-64f5779d4ff0	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
831b1067-ca80-4b84-a2e2-cf55e4f25a7a	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
7c50dc46-5c98-4a86-9c98-b405a8af4646	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
6934a4c9-f355-402f-8a74-82fcb3404822	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
5aa25615-a522-4c00-97ad-addac1bcc2a8	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
41be9c6b-8f5c-45a1-b803-4e6be9466666	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
a5d89986-969d-49a0-9825-6bf2ab4e5cb9	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
b6ecbcb0-92ab-4c13-9af7-9cc7257fc9ea	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
dd2162c8-ce78-4dda-9998-defa2f8bb38e	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
fdb60914-3901-4673-a1fb-fee9cb3e8e4b	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
01238ace-504b-4296-acda-a2a6964b6502	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
1397d897-38c3-4a60-9c98-08b16bcac66d	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
73d8dabb-5731-431b-b98f-8b0425cb6b74	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
af86ee30-9489-4942-bd99-e62b81d9428f	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
49b68275-abe6-4107-b3af-ae13cc689ee7	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
7179b9db-f455-4622-a5fc-503a48d4e0c8	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
02285cba-0554-442c-9f97-9f0affdfc6cc	ru	Скоро будет	2	import	2026-01-17 22:30:43.26705+00
3debf8c8-e80f-42d2-9bc4-2d8950de8e06	ru	Скоро будет	2	import	2026-01-17 22:30:43.299741+00
d2659d23-21ab-4664-a710-a5b54fabe171	ru	Скоро будет	2	import	2026-01-17 22:30:43.301742+00
141781f2-30e0-4ae3-9de1-ba03add658cc	ru	Скоро будет	2	import	2026-01-17 22:30:43.301742+00
438a5ce3-fc80-4645-b55a-33d5a869e901	ru	Скоро будет	2	import	2026-01-17 22:30:43.301742+00
4626446d-04a1-4fcb-aba1-f9631092d77f	ru	Скоро будет	2	import	2026-01-17 22:30:43.301742+00
c9a927ef-0c56-479b-ab08-f6af39f3eee1	ru	Скоро будет	2	import	2026-01-17 22:30:43.304163+00
d7ff9af0-e625-4052-b76a-307f5de31cf3	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
19804fd5-3ee7-4b28-8349-8cf3a343e332	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
725c4251-57b6-487e-a5ac-e69c5b9c5ce2	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
69f680e1-fbea-4159-a3a7-d60ce7ccacb8	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
b9791e58-ccd7-4139-aabb-b65b34464b84	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
f2487620-6447-4d20-b562-207ba31f726e	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
262b5ec2-b1ae-4d88-b118-af8d8bd64f74	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
7137ba05-6341-49fb-8cb9-2ad48f549119	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
ebe253d3-79ee-4ede-b0b7-308861a27493	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
f07a37b7-c928-4db9-b329-37691e8aeb1f	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
5b405fdb-9e8a-4d54-be43-5a1b8cf1163f	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
8053ecba-9a23-450b-93fc-ff6fc86b7c98	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
5e9826f2-0f17-42a9-93fa-cf0c7143a5af	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
75ffc6dd-1503-4537-8a98-21d9cb4865d6	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
2b9496be-6de3-4012-a0e8-b3c2d27510a1	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
4f6cc43d-c254-46c0-889e-631654249f9a	ru	Скоро будет	2	import	2026-01-17 22:30:43.306071+00
33f5fc91-cdc1-4949-87a6-514f956e8166	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
40d1612e-b509-44ff-917a-ea660d59c95b	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
773f2b42-b216-4905-bf9b-2f16655b57f8	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
110dac52-1b61-49fb-9300-73b994166f35	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
eca78d27-6cf3-42c3-86f9-500eab275dda	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
9df2edc1-30e3-4a3f-9f01-bbb37aaa9880	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
4262ada4-ab26-4e80-9353-c606d5026f90	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
35d798eb-d873-4529-a188-f0ff7845f28d	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
30498bc1-8049-4b9d-b0d4-8e063dc38616	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
528ad84a-4cd6-4e7d-a9fe-a82929a0013f	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
3c78b094-e31a-4626-a4b3-1c304095876c	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
39bc8f9f-5ecd-49d6-8394-a90cc5028ecb	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
c2310b0c-2a64-47b2-8ab5-de99fce64835	ru	Скоро будет	2	import	2026-01-17 22:30:43.310948+00
55501f25-6ada-4572-88cc-3ff40dfb3bf8	ru	Скоро будет	2	import	2026-01-17 22:30:43.315092+00
2cad38a9-0693-4ddf-adb0-1faddc102cde	ru	Скоро будет	2	import	2026-01-17 22:30:43.315092+00
e278da47-a718-4dd9-8169-c04328cbd577	ru	Скоро будет	2	import	2026-01-17 22:30:43.315092+00
11895cf8-6862-4601-97ac-c57e29455477	ru	Скоро будет	2	import	2026-01-17 22:30:43.315092+00
\.


--
-- Data for Name: chapters; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.chapters (id, novel_id, number, slug, title, views, published_at, created_at, updated_at) FROM stdin;
7427c8b3-d65d-407a-b8a2-b2400a93183a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	1.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
87a006b4-edb9-425b-9d9c-57de8a18ab5f	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	2.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
9bb21178-f58d-4821-bf1b-10a23e175949	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	3.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
29bd06c1-7c59-44f6-8293-c06ec06f480f	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	4.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
c27f3c04-4436-4fc2-b0c8-6cf169a2a87d	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	5.00	\N	Том 1 Глава 5	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
93941aa7-b3cf-4b0f-a0ed-0f76bbba2f12	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	6.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
31a2686c-6723-4ee6-8bde-834048f20990	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	7.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
b535bbee-5632-4c24-bd14-d84c52bdb29e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	8.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
ca7db5ff-526e-4106-a1d6-16215c738054	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	9.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
1b1426b2-c8ab-455c-a8b9-8aa7f70cda06	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	10.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
8b11fc6a-25db-4d9a-9c80-8060828f983e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	11.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
8bda608d-9b30-4a7f-9adb-3cf4cbd84b13	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	12.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
ac5da8cd-bf46-49fa-9ee5-9e4b2efacca5	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	13.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
1bffb6b2-26d1-42e6-9a22-e7870091bc94	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	14.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
5fa92323-a25c-4413-9b73-07c83d732952	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	15.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
3c1266e1-1f79-4510-994e-43a6a77a3bf7	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	16.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
5e940078-ea05-4602-a57d-5aeda861d035	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	17.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
964cbcd7-be1c-4efb-a290-c7ca0f0859ed	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	18.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
c1a36b57-6ed2-4cf1-8760-bb8d27dbadb1	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	19.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
9516f965-36a6-4cb4-b665-5f2c984d721e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	20.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
857430f8-b9cd-4105-a4da-e749e66ed09a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	21.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
3687251c-efd4-40a5-9b3e-4cf98ad2bf4c	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	22.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
2ffcc8c5-14b4-46a8-9cd9-676b7082ce06	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	23.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
25088950-e2f8-4028-bd98-0e5f116cbb3d	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	24.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
31c50974-7e2d-4ba9-9dd5-909647491016	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	25.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
3eae23d3-eb46-46f8-8ed5-c6e3d6b2c00a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	26.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
cc28095d-723e-48e7-bf84-47661443dc59	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	27.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
9948543a-3338-4d0e-8ec7-34a0713b39f1	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	28.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
dc5573a7-accb-4d69-8df6-b848f38b487c	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	29.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
e15fcca1-ae5c-45c3-ab52-1699ac4e3b3c	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	30.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
cf9b1825-fdd5-4a21-80ec-d1eea8547782	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	31.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
c0c6aff3-9a5e-4826-ae59-3afc3eceb5ed	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	32.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
f62a7132-a58e-4ebc-9b11-7281803c2562	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	33.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
dc1b189d-b2ae-4af9-bb86-3160373af705	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	34.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
2a05c44c-7664-43bd-8836-d199ec58379a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	35.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
b06ff097-18f9-45d3-9fbd-9392d3bbf8f7	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	36.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
13cac42f-efca-41e2-be2c-c8714e7f748b	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	37.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
eba42a24-906f-46c9-918e-a9a07203b997	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	38.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
e9b0c65d-2d27-4c9e-a4a1-a248d2e1463e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	39.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
2c546d05-11ef-4a22-bef1-43c87139bc26	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	40.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
3ce08b80-2557-48ee-ad09-263e16cfccad	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	41.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
6bbb1826-6488-4ceb-b91b-c21e20f21b4a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	42.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
0227ad5c-a8f5-4b29-a982-df3188f9b4eb	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	43.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
077644e0-20a2-4e97-9ac9-71708670dd4c	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	44.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
b1721dd0-8679-40a4-85a0-de4690136200	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	45.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
050f1e13-9af7-447c-9f4a-a9b6859c97a0	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	46.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
15cead81-2436-4638-80dd-7665bef242ca	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	47.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
64ed73a6-ef3f-4c2e-8508-04c0762a6114	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	48.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
7a5b9278-a963-472c-936b-e6b9ad9a3d92	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	49.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
79ce09e5-164f-4b3a-b702-d8258e74a3b9	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	50.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
cf82adda-1a6c-407a-8b17-0613077e6f26	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	51.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
a80b5117-751d-41eb-894e-a6fc941c2a9a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	52.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
43afbd68-5759-4a5a-84e6-acb0165f2dc6	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	53.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
3059c97e-ed04-4535-8dc9-e2e2081273a7	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	54.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
262602d2-e610-47e0-97b9-651b61a081c6	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	55.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
8f88a42e-930f-43cf-8826-2682f4626393	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	56.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
4bb71bb7-9aeb-4d0e-ab14-7a9df1f72679	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	57.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
9aff976f-6566-4dc7-b929-ee7ea2de891a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	58.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
bcdced7c-c60b-4ac5-90e6-142ea8069c60	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	59.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
c62858e8-3394-4292-a6c1-508842c5ca45	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	60.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
772b6122-3540-4e37-b7dd-35fc1ab177c0	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	61.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
5e7b67d5-98ce-4d13-8bca-c20012e692c9	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	62.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
64856e08-d899-46e8-901e-c76af22d9e65	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	63.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
268692ad-f1e9-4e54-8963-36ef12b5262f	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	64.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
7380d62c-1fa0-4516-be36-e794f7f8f73c	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	65.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
12f52664-5a3a-4465-8a45-1650d71f69eb	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	66.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
f8f19ddf-a8fa-43fe-bb8b-00b5adfc4922	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	67.00	\N	Начало 2 сезона	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
89f6d36a-71c9-4e10-b5c8-6b323712bf81	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	68.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
fd24295e-d635-49ad-a28a-928ea5cbb5ee	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	69.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
61c5a2e3-94f8-4460-a80d-cc29aa1197f5	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	70.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
6e46fd81-ecfe-4759-9757-48028041d27b	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	71.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
e7650081-c532-47eb-a9f5-b0f31bb50319	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	72.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
c55e7ece-7bac-428f-9b35-036684eed7ae	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	73.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
35e7f90c-a17a-4280-b803-1b56d3fbd33e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	74.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
beb21418-fa55-4af2-b2e2-8c83f4bca26e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	75.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
eede3327-1a63-4d17-957c-2c30423415d4	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	76.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
e40edea1-5280-48eb-ac4e-ab626034a442	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	77.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
aa5cd215-c1de-4b9c-aaf5-fdcd71344e10	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	78.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
41a11632-cf61-462b-a7e8-f9dcfa6334e3	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	79.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
63872525-4b07-4f5a-a2f3-64f5779d4ff0	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	80.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
831b1067-ca80-4b84-a2e2-cf55e4f25a7a	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	81.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
7c50dc46-5c98-4a86-9c98-b405a8af4646	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	82.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
6934a4c9-f355-402f-8a74-82fcb3404822	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	83.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
5aa25615-a522-4c00-97ad-addac1bcc2a8	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	84.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
41be9c6b-8f5c-45a1-b803-4e6be9466666	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	85.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
a5d89986-969d-49a0-9825-6bf2ab4e5cb9	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	86.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
b6ecbcb0-92ab-4c13-9af7-9cc7257fc9ea	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	87.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
dd2162c8-ce78-4dda-9998-defa2f8bb38e	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	88.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
fdb60914-3901-4673-a1fb-fee9cb3e8e4b	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	89.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
01238ace-504b-4296-acda-a2a6964b6502	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	90.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
1397d897-38c3-4a60-9c98-08b16bcac66d	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	91.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
73d8dabb-5731-431b-b98f-8b0425cb6b74	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	92.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
af86ee30-9489-4942-bd99-e62b81d9428f	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	93.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
49b68275-abe6-4107-b3af-ae13cc689ee7	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	94.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
7179b9db-f455-4622-a5fc-503a48d4e0c8	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	95.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
02285cba-0554-442c-9f97-9f0affdfc6cc	72cd95ba-f0f0-4330-98a4-f171c7f8cabe	96.00	\N	\N	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
4626446d-04a1-4fcb-aba1-f9631092d77f	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	4.00	\N	\N	0	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00
c9a927ef-0c56-479b-ab08-f6af39f3eee1	a239b948-358f-4179-808d-f8c6bb5715c3	1.00	\N	Ваншот для X конкурса	0	2026-01-17 22:30:43.304163+00	2026-01-17 22:30:43.304163+00	2026-01-17 22:30:43.304163+00
d7ff9af0-e625-4052-b76a-307f5de31cf3	92a1b7fa-070f-4c56-8e26-afc1db52130c	1.00	\N	Номер (1)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
19804fd5-3ee7-4b28-8349-8cf3a343e332	92a1b7fa-070f-4c56-8e26-afc1db52130c	2.00	\N	Номер (2)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
b9791e58-ccd7-4139-aabb-b65b34464b84	92a1b7fa-070f-4c56-8e26-afc1db52130c	5.00	\N	Наследие убийцы (2)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
f2487620-6447-4d20-b562-207ba31f726e	92a1b7fa-070f-4c56-8e26-afc1db52130c	6.00	\N	Наследие убийцы (3)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
262b5ec2-b1ae-4d88-b118-af8d8bd64f74	92a1b7fa-070f-4c56-8e26-afc1db52130c	7.00	\N	Наследие Джи Хён (4)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
7137ba05-6341-49fb-8cb9-2ad48f549119	92a1b7fa-070f-4c56-8e26-afc1db52130c	8.00	\N	Дьявольское доказательство (1)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
ebe253d3-79ee-4ede-b0b7-308861a27493	92a1b7fa-070f-4c56-8e26-afc1db52130c	9.00	\N	Дьявольское доказательство (2)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
f07a37b7-c928-4db9-b329-37691e8aeb1f	92a1b7fa-070f-4c56-8e26-afc1db52130c	10.00	\N	Двойная слежка (1)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
5b405fdb-9e8a-4d54-be43-5a1b8cf1163f	92a1b7fa-070f-4c56-8e26-afc1db52130c	11.00	\N	Двойная слежка (2)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
8053ecba-9a23-450b-93fc-ff6fc86b7c98	92a1b7fa-070f-4c56-8e26-afc1db52130c	12.00	\N	Двойная слежка (3)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
5e9826f2-0f17-42a9-93fa-cf0c7143a5af	92a1b7fa-070f-4c56-8e26-afc1db52130c	13.00	\N	Двойная слежка (4)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
75ffc6dd-1503-4537-8a98-21d9cb4865d6	92a1b7fa-070f-4c56-8e26-afc1db52130c	14.00	\N	Право на последнее слово	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
2b9496be-6de3-4012-a0e8-b3c2d27510a1	92a1b7fa-070f-4c56-8e26-afc1db52130c	15.00	\N	Наглая ложь (1)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
4f6cc43d-c254-46c0-889e-631654249f9a	92a1b7fa-070f-4c56-8e26-afc1db52130c	16.00	\N	Наглая ложь (2)	0	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
33f5fc91-cdc1-4949-87a6-514f956e8166	e69f0077-8bc6-47a2-9fba-6fa47a926984	1.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
40d1612e-b509-44ff-917a-ea660d59c95b	e69f0077-8bc6-47a2-9fba-6fa47a926984	2.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
773f2b42-b216-4905-bf9b-2f16655b57f8	e69f0077-8bc6-47a2-9fba-6fa47a926984	3.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
110dac52-1b61-49fb-9300-73b994166f35	e69f0077-8bc6-47a2-9fba-6fa47a926984	4.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
eca78d27-6cf3-42c3-86f9-500eab275dda	e69f0077-8bc6-47a2-9fba-6fa47a926984	5.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
9df2edc1-30e3-4a3f-9f01-bbb37aaa9880	e69f0077-8bc6-47a2-9fba-6fa47a926984	6.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
4262ada4-ab26-4e80-9353-c606d5026f90	e69f0077-8bc6-47a2-9fba-6fa47a926984	7.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
35d798eb-d873-4529-a188-f0ff7845f28d	e69f0077-8bc6-47a2-9fba-6fa47a926984	8.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
30498bc1-8049-4b9d-b0d4-8e063dc38616	e69f0077-8bc6-47a2-9fba-6fa47a926984	9.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
528ad84a-4cd6-4e7d-a9fe-a82929a0013f	e69f0077-8bc6-47a2-9fba-6fa47a926984	10.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
3c78b094-e31a-4626-a4b3-1c304095876c	e69f0077-8bc6-47a2-9fba-6fa47a926984	11.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
39bc8f9f-5ecd-49d6-8394-a90cc5028ecb	e69f0077-8bc6-47a2-9fba-6fa47a926984	12.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
c2310b0c-2a64-47b2-8ab5-de99fce64835	e69f0077-8bc6-47a2-9fba-6fa47a926984	13.00	\N	\N	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
55501f25-6ada-4572-88cc-3ff40dfb3bf8	d15e5d69-6582-4ac9-97cf-997c49b90aff	1.00	\N	Хатори и Фурута	0	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
2cad38a9-0693-4ddf-adb0-1faddc102cde	d15e5d69-6582-4ac9-97cf-997c49b90aff	2.00	\N	Тот, кого я ждала...	0	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
e278da47-a718-4dd9-8169-c04328cbd577	d15e5d69-6582-4ac9-97cf-997c49b90aff	3.00	\N	Рецепт весёлого дня	0	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
11895cf8-6862-4601-97ac-c57e29455477	d15e5d69-6582-4ac9-97cf-997c49b90aff	4.00	\N	Существо зазеркалья	0	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
725c4251-57b6-487e-a5ac-e69c5b9c5ce2	92a1b7fa-070f-4c56-8e26-afc1db52130c	3.00	\N	Мерцание, предотвращение рецидива	1	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-22 21:00:55.295666+00
141781f2-30e0-4ae3-9de1-ba03add658cc	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	2.00	\N	\N	1	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00	2026-01-21 19:54:49.337486+00
438a5ce3-fc80-4645-b55a-33d5a869e901	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	3.00	\N	\N	1	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00	2026-01-21 19:54:52.481353+00
3debf8c8-e80f-42d2-9bc4-2d8950de8e06	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	1.00	\N	Cклонность к неудачам	4	2026-01-17 22:30:43.299741+00	2026-01-17 22:30:43.299741+00	2026-01-22 20:34:54.662759+00
69f680e1-fbea-4159-a3a7-d60ce7ccacb8	92a1b7fa-070f-4c56-8e26-afc1db52130c	4.00	\N	Наследие убийцы (1)	1	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00	2026-01-22 21:00:59.279132+00
d2659d23-21ab-4664-a710-a5b54fabe171	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	1.00	\N	\N	1	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00	2026-01-22 21:17:01.194436+00
\.


--
-- Data for Name: collection_items; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.collection_items (id, collection_id, novel_id, "position", note, added_at) FROM stdin;
\.


--
-- Data for Name: collection_votes; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.collection_votes (id, collection_id, user_id, value, created_at) FROM stdin;
\.


--
-- Data for Name: collections; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.collections (id, user_id, slug, title, description, cover_url, is_public, is_featured, views_count, votes_count, items_count, created_at, updated_at) FROM stdin;
01b938fd-3b18-42e1-a5aa-6a9ae3f96fcf	ce76495c-3f80-4c7a-be56-3d5994b49b03	iachs-iaa	ячс яа	az fsadf sdaf sad fsad fsadfasdf sad fsdf	https://remanga.org/media/titles/admirals-monster-wife/cover_92ea0590e99343b5.webp	t	f	0	0	0	2026-01-16 23:57:22.788492+00	2026-01-16 23:57:22.788492+00
13528034-12ec-4b5f-9f63-6da4e1da44f9	ce76495c-3f80-4c7a-be56-3d5994b49b03	zxc	zxc	qwer weagsadg asdg	https://remanga.org/media/titles/hiddn-sint/612004ad6997758d5b35b4d16f656508.jpg	t	f	0	0	0	2026-01-17 00:10:05.990411+00	2026-01-17 00:10:05.990411+00
c5fca40f-4813-4bd8-8536-7a3e32eff63f	ce76495c-3f80-4c7a-be56-3d5994b49b03	zxc-234qqrefsadfzsdf	zxc 234qqrefsadfzsdf	df gt354wt wergfvsdgdsfg dsfg sdgf	https://remanga.org/media/titles/a-became-my-heroines-wife/cover_885ae8367e304f73.webp	t	f	0	0	0	2026-01-17 00:12:32.484784+00	2026-01-17 00:12:32.484784+00
\.


--
-- Data for Name: comment_reports; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.comment_reports (id, comment_id, user_id, reason, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: comment_votes; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.comment_votes (comment_id, user_id, value, created_at) FROM stdin;
3b68fa60-e47e-416b-a308-d504a7b5bec3	ce76495c-3f80-4c7a-be56-3d5994b49b03	1	2026-01-21 22:08:13.07167+00
1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	ce76495c-3f80-4c7a-be56-3d5994b49b03	-1	2026-01-21 22:11:35.368897+00
59af45e7-1a76-4c0f-bca8-8ca91030a349	ce76495c-3f80-4c7a-be56-3d5994b49b03	-1	2026-01-21 22:11:46.310086+00
d8c53acf-7200-413f-a9e3-478fcd590b50	ce76495c-3f80-4c7a-be56-3d5994b49b03	-1	2026-01-21 22:12:00.916181+00
c6d79606-ea22-483b-ab72-2c6740adc300	ce76495c-3f80-4c7a-be56-3d5994b49b03	-1	2026-01-22 20:20:46.396502+00
8b389d1f-d0ac-4a3a-8883-e8dc372a7038	ce76495c-3f80-4c7a-be56-3d5994b49b03	1	2026-01-22 20:20:48.393156+00
25d1d476-e49a-4412-af42-3c030651693e	ce76495c-3f80-4c7a-be56-3d5994b49b03	-1	2026-01-22 20:35:02.749669+00
\.


--
-- Data for Name: comments; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.comments (id, novel_id, chapter_id, user_id, parent_id, content, is_spoiler, likes_count, dislikes_count, replies_count, created_at, updated_at, deleted_at, target_type, target_id, body, is_deleted, root_id, depth, anchor) FROM stdin;
222dc36d-e04c-4eb6-ba64-04fca8216eb5	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	[удалено]	f	0	0	0	2026-01-21 22:00:22.018236+00	2026-01-21 22:02:36.888177+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	[удалено]	t	222dc36d-e04c-4eb6-ba64-04fca8216eb5	0	\N
7c9c3a75-91be-4648-b180-5b2be0fa400e	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	[удалено]	t	0	0	0	2026-01-21 22:00:13.42024+00	2026-01-21 22:02:45.030085+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	[удалено]	t	7c9c3a75-91be-4648-b180-5b2be0fa400e	0	\N
59af45e7-1a76-4c0f-bca8-8ca91030a349	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	ыфваывфа	f	0	1	0	2026-01-21 22:11:38.556963+00	2026-01-21 22:11:38.556963+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	ыфваывфа	f	1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	1	\N
c8355372-d521-43b0-963e-ad3ea04d5800	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	d8c53acf-7200-413f-a9e3-478fcd590b50	123	f	0	0	0	2026-01-21 22:12:04.139627+00	2026-01-21 22:12:04.139627+00	\N	chapter	3debf8c8-e80f-42d2-9bc4-2d8950de8e06	123	f	d8c53acf-7200-413f-a9e3-478fcd590b50	1	\N
d8c53acf-7200-413f-a9e3-478fcd590b50	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	фывавфыа	t	0	1	1	2026-01-21 22:11:58.485399+00	2026-01-21 22:11:58.485399+00	\N	chapter	3debf8c8-e80f-42d2-9bc4-2d8950de8e06	фывавфыа	f	d8c53acf-7200-413f-a9e3-478fcd590b50	0	\N
aa6288ff-edae-45cd-8f6d-2e2a7577c06b	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	ывыфвфыв	f	0	0	1	2026-01-21 22:02:58.886263+00	2026-01-21 22:02:58.886263+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	ывыфвфыв	f	aa6288ff-edae-45cd-8f6d-2e2a7577c06b	0	\N
c6d79606-ea22-483b-ab72-2c6740adc300	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	8b389d1f-d0ac-4a3a-8883-e8dc372a7038	sdfdfs	f	0	1	0	2026-01-22 20:20:43.251096+00	2026-01-22 20:20:43.251096+00	\N	profile	ce76495c-3f80-4c7a-be56-3d5994b49b03	sdfdfs	f	8b389d1f-d0ac-4a3a-8883-e8dc372a7038	1	\N
8b389d1f-d0ac-4a3a-8883-e8dc372a7038	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	asfsad fsdf 	t	1	0	1	2026-01-22 20:20:39.025813+00	2026-01-22 20:20:39.025813+00	\N	profile	ce76495c-3f80-4c7a-be56-3d5994b49b03	asfsad fsdf 	f	8b389d1f-d0ac-4a3a-8883-e8dc372a7038	0	\N
5db7d6cf-ed3a-40e2-94ce-525270df8682	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	ddd	f	0	0	0	2026-01-22 20:22:05.126509+00	2026-01-22 20:22:05.126509+00	\N	novel	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	ddd	f	5db7d6cf-ed3a-40e2-94ce-525270df8682	0	\N
3b68fa60-e47e-416b-a308-d504a7b5bec3	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	aa6288ff-edae-45cd-8f6d-2e2a7577c06b	123	f	1	1	0	2026-01-21 22:03:09.380944+00	2026-01-21 22:03:09.380944+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	123	f	aa6288ff-edae-45cd-8f6d-2e2a7577c06b	1	\N
25d1d476-e49a-4412-af42-3c030651693e	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	123cadfg dsafg	t	0	1	0	2026-01-22 20:35:00.827414+00	2026-01-22 20:35:00.827414+00	\N	chapter	3debf8c8-e80f-42d2-9bc4-2d8950de8e06	123cadfg dsafg	f	25d1d476-e49a-4412-af42-3c030651693e	0	\N
1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	\N	\N	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	ввв	f	0	1	1	2026-01-21 22:11:33.96062+00	2026-01-21 22:11:33.96062+00	\N	novel	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	ввв	f	1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	0	\N
\.


--
-- Data for Name: daily_vote_grants; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.daily_vote_grants (id, grant_date, users_processed, total_votes_granted, started_at, completed_at, status, error_message) FROM stdin;
e544493d-7cfc-48d2-bd5c-418be1909367	2026-01-21	2	2	2026-01-21 21:30:11.534939+00	2026-01-21 21:30:11.645438+00	completed	\N
\.


--
-- Data for Name: genre_localizations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.genre_localizations (genre_id, lang, name) FROM stdin;
\.


--
-- Data for Name: genres; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.genres (id, slug, created_at) FROM stdin;
\.


--
-- Data for Name: leaderboard_cache; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.leaderboard_cache (id, period, user_id, tickets_spent, rank, calculated_at) FROM stdin;
\.


--
-- Data for Name: news_localizations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.news_localizations (id, news_id, lang, title, summary, content, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: news_posts; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.news_posts (id, slug, title, summary, content, cover_url, category, author_id, is_published, is_pinned, views_count, comments_count, published_at, created_at, updated_at) FROM stdin;
22d06f06-2be2-4743-a0c6-ff237ce4e450	iachs1234	ячс1234	 ывапыв фпыфвам	zxcvcwer tgfsdgsdg q34wegf		announcement	ce76495c-3f80-4c7a-be56-3d5994b49b03	t	f	0	0	2026-01-16 23:24:12.619135+00	2026-01-16 23:24:12.617592+00	2026-01-16 23:24:12.619234+00
\.


--
-- Data for Name: novel_authors; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_authors (novel_id, author_id, is_primary, sort_order, created_at) FROM stdin;
\.


--
-- Data for Name: novel_edit_history; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_edit_history (id, novel_id, request_id, user_id, field_type, lang, old_value, new_value, created_at) FROM stdin;
\.


--
-- Data for Name: novel_edit_request_changes; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_edit_request_changes (id, request_id, field_type, lang, old_value, new_value, created_at) FROM stdin;
\.


--
-- Data for Name: novel_edit_requests; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_edit_requests (id, novel_id, user_id, status, edit_reason, moderator_id, moderator_comment, reviewed_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: novel_genres; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_genres (novel_id, genre_id) FROM stdin;
\.


--
-- Data for Name: novel_localizations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_localizations (novel_id, lang, title, description, alt_titles, search_vector, created_at, updated_at) FROM stdin;
6db455a3-5bf9-4e65-a07a-0f6f3118a452	ru	Сестры близняшки не дают покоя Нэдзуми	\N	{}	'близняшки':2A 'дают':4A 'не':3A 'нэдзуми':6A 'покоя':5A 'сестры':1A	2026-01-17 22:24:28.080783+00	2026-01-17 22:24:28.080783+00
5be035c3-ec48-445d-88e1-b08c1b6f1bb2	ru	Карл – обходчик подземелий	\N	{}	'карл':1A 'обходчик':2A 'подземелий':3A	2026-01-17 22:24:28.089772+00	2026-01-17 22:24:28.089772+00
b5818047-3d8a-4700-af9a-1758c1db91c5	ru	Первый	\N	{}	'первый':1A	2026-01-17 22:24:28.091236+00	2026-01-17 22:24:28.091236+00
bd31a32c-1889-4016-80df-631d232bdc7f	ru	Тайная семья	\N	{}	'семья':2A 'тайная':1A	2026-01-17 22:24:28.092656+00	2026-01-17 22:24:28.092656+00
6d8c6389-1ca6-4c9f-8212-1c94b880546d	ru	Странная, но эффективная жизнь злодейки	\N	{}	'жизнь':4A 'злодейки':5A 'но':2A 'странная':1A 'эффективная':3A	2026-01-17 22:24:28.09436+00	2026-01-17 22:24:28.09436+00
da4defbb-aef7-4b9b-a1a4-ec6f96172bc0	ru	Моя фальшивая любовь	\N	{}	'любовь':3A 'моя':1A 'фальшивая':2A	2026-01-17 22:24:28.095973+00	2026-01-17 22:24:28.095973+00
f0b74a56-9c98-4354-87db-45e7808e271e	ru	Свидетель моей любви	\N	{}	'любви':3A 'моей':2A 'свидетель':1A	2026-01-17 22:24:28.097565+00	2026-01-17 22:24:28.097565+00
7c28c91b-ebdd-4407-b3e9-c74e7c9ff980	ru	Коты-воители: Путь пророчеств	\N	{}	'воители':3A 'коты':2A 'коты-воители':1A 'пророчеств':5A 'путь':4A	2026-01-17 22:24:28.098915+00	2026-01-17 22:24:28.098915+00
0097c045-1941-4706-adb9-b1b592200c7e	ru	Фальшивый мастер, который случайно стал самым сильным	\N	{}	'который':3A 'мастер':2A 'самым':6A 'сильным':7A 'случайно':4A 'стал':5A 'фальшивый':1A	2026-01-17 22:24:28.100432+00	2026-01-17 22:24:28.100432+00
1b3a53fc-59ff-46ae-ab60-4b9f50e3ae19	ru	Все ли матери обречены на смерть?	\N	{}	'все':1A 'ли':2A 'матери':3A 'на':5A 'обречены':4A 'смерть':6A	2026-01-17 22:24:28.101758+00	2026-01-17 22:24:28.101758+00
a4b9d9d2-1ac0-40e7-9a22-57d5c2addf6f	ru	Безнравственный скандал	\N	{}	'безнравственный':1A 'скандал':2A	2026-01-17 22:24:28.103183+00	2026-01-17 22:24:28.103183+00
c66807fe-8d42-4290-8bbc-f0a8bf807c3a	ru	Я хочу съесть твои кишки	\N	{}	'кишки':5A 'съесть':3A 'твои':4A 'хочу':2A 'я':1A	2026-01-17 22:24:28.104482+00	2026-01-17 22:24:28.104482+00
8e9c621f-c81d-43ca-95a5-b662cdb79516	ru	Одержимая злодейка сеет хаос	\N	{}	'злодейка':2A 'одержимая':1A 'сеет':3A 'хаос':4A	2026-01-17 22:24:28.105791+00	2026-01-17 22:24:28.105791+00
72cd95ba-f0f0-4330-98a4-f171c7f8cabe	ru	Бастард был императором	Иан, маг империи Бариэль, в юном возрасте стал императором, однако из-за мятежа, поднятого его племянником, оказался в заключении. \n\nС помощью запретной магии его спасла волшебница Наум. Это привело к тому, что Иан оказался в теле незаконнорожденного сына маркграфа, чей род угас 100 лет назад. \n \nСына маркграфа вот-вот возьмут в заложники варвары. Иану предстоит погрузиться в хаос заговоров и интриг!	{"Margrave's Bastard Son Was the Emperor","Внебрачный сын маркграфа оказался императором",边境庶子是皇帝,辺境伯家の落ちこぼれは皇帝だった,"변경백 서자는 황제였다"}	'100':47B 'bastard':69A 'emperor':73A 'margrave':67A 's':68A 'son':70A 'the':72A 'was':71A 'бариэль':7B 'бастард':1A 'был':2A 'в':8B,22B,39B,56B,62B 'варвары':58B 'внебрачный':74A 'возрасте':10B 'возьмут':55B 'волшебница':30B 'вот':53B,54B 'вот-вот':52B 'его':19B,28B 'за':16B 'заговоров':64B 'заключении':23B 'заложники':57B 'запретной':26B 'и':65B 'иан':4B,37B 'иану':59B 'из':15B 'из-за':14B 'императором':3A,12B,78A 'империи':6B 'интриг':66B 'к':34B 'лет':48B 'маг':5B 'магии':27B 'маркграфа':43B,51B,76A 'мятежа':17B 'назад':49B 'наум':31B 'незаконнорожденного':41B 'однако':13B 'оказался':21B,38B,77A 'племянником':20B 'погрузиться':61B 'поднятого':18B 'помощью':25B 'предстоит':60B 'привело':33B 'род':45B 'с':24B 'спасла':29B 'стал':11B 'сын':75A 'сына':42B,50B 'теле':40B 'тому':35B 'угас':46B 'хаос':63B 'чей':44B 'что':36B 'это':32B 'юном':9B '边境庶子是皇帝':79A '辺境伯家の落ちこぼれは皇帝だった':80A '변경백':81A '서자는':82A '황제였다':83A	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	ru	Неудачное свидание	Неудачливая по жизни Хайда в порыве отчаяния регистрируется в приложении для знакомств. Там она знакомится с Курои - мужчиной, который зарегистрировался по той же причине. С самого начала разговор у них не клеится, и между ними повисает гнетуще унылая атмосфера. Что бы они ни делали, всё выходит боком, и свидание получается на удивление неинтересным и провальным. «С таким настроением второго раза точно не будет» - подумала Хайда… \nНо постойте, почему же они снова договариваются о встрече?!	{盛りあがらないデート}	'атмосфера':41B 'боком':49B 'будет':65B 'бы':43B 'в':7B,11B 'встрече':76B 'всё':47B 'второго':61B 'выходит':48B 'гнетуще':39B 'делали':46B 'для':13B 'договариваются':74B 'же':25B,71B 'жизни':5B 'зарегистрировался':22B 'знакомится':17B 'знакомств':14B 'и':35B,50B,56B 'клеится':34B 'который':21B 'курои':19B 'между':36B 'мужчиной':20B 'на':53B 'настроением':60B 'начала':29B 'не':33B,64B 'неинтересным':55B 'неудачливая':3B 'неудачное':1A 'ни':45B 'ними':37B 'них':32B 'но':68B 'о':75B 'она':16B 'они':44B,72B 'отчаяния':9B 'по':4B,23B 'повисает':38B 'подумала':66B 'получается':52B 'порыве':8B 'постойте':69B 'почему':70B 'приложении':12B 'причине':26B 'провальным':57B 'раза':62B 'разговор':30B 'регистрируется':10B 'с':18B,27B,58B 'самого':28B 'свидание':2A,51B 'снова':73B 'таким':59B 'там':15B 'той':24B 'точно':63B 'у':31B 'удивление':54B 'унылая':40B 'хайда':6B,67B 'что':42B '盛りあがらないデート':77A	2026-01-17 22:30:43.299741+00	2026-01-17 22:30:43.299741+00
6dad104f-6d7e-4c26-927e-e4457ab8fb1b	ru	Я тебя поймал	«Спасибо, что стал моей последней удачей!»\nСентябрь 2015 года. Абитуриент Квон Мин, входящий в топ 0,1% лучших учеников, случайно раскрывает личность загадочной Ча Юбин — ученицы из параллельного класса, которая была для него лишь иллюзией, забравшей двадцать минут на пробном экзамене.\nТеперь их двое. Каждый день у них есть лишь по тридцать минут, чтобы распутать самую сложную из всех возможных головоломок — тайну собственных удачи и судьбы.\nВсё это — их юность, что состоит из мимолётных, сияющих мгновений, которые так хочется удержать. Мгновений, которые даруются даже девятнадцатилетним, чья беззаботная жизнь стремительно несётся к финалу под звуки обратного отсчёта до экзаменов.	{아이캣치유}	'0':19B '1':20B '2015':11B 'абитуриент':13B 'беззаботная':90B 'была':34B 'в':17B 'возможных':63B 'всех':62B 'всё':70B 'входящий':16B 'года':12B 'головоломок':64B 'даже':87B 'даруются':86B 'двадцать':40B 'двое':47B 'девятнадцатилетним':88B 'день':49B 'для':35B 'до':100B 'есть':52B 'жизнь':91B 'забравшей':39B 'загадочной':26B 'звуки':97B 'и':68B 'из':30B,61B,76B 'иллюзией':38B 'их':46B,72B 'к':94B 'каждый':48B 'квон':14B 'класса':32B 'которая':33B 'которые':80B,85B 'личность':25B 'лишь':37B,53B 'лучших':21B 'мгновений':79B,84B 'мимолётных':77B 'мин':15B 'минут':41B,56B 'моей':7B 'на':42B 'него':36B 'несётся':93B 'них':51B 'обратного':98B 'отсчёта':99B 'параллельного':31B 'по':54B 'под':96B 'поймал':3A 'последней':8B 'пробном':43B 'раскрывает':24B 'распутать':58B 'самую':59B 'сентябрь':10B 'сияющих':78B 'сложную':60B 'случайно':23B 'собственных':66B 'состоит':75B 'спасибо':4B 'стал':6B 'стремительно':92B 'судьбы':69B 'тайну':65B 'так':81B 'тебя':2A 'теперь':45B 'топ':18B 'тридцать':55B 'у':50B 'удачей':9B 'удачи':67B 'удержать':83B 'учеников':22B 'ученицы':29B 'финалу':95B 'хочется':82B 'ча':27B 'что':5B,74B 'чтобы':57B 'чья':89B 'экзамене':44B 'экзаменов':101B 'это':71B 'юбин':28B 'юность':73B 'я':1A '아이캣치유':102A	2026-01-17 22:30:43.301742+00	2026-01-17 22:30:43.301742+00
a239b948-358f-4179-808d-f8c6bb5715c3	ru	А что, если я накормлю молчаливую гяру, сидящую рядом со мной?	В средней школе надо мной постоянно издевались гяру. В старшей школе я решил изменить свою жизнь, но за соседней партой оказалась гяру! Но, кажется, она немного отличается от других гяру...	{"Если я накормлю молчаливую девушку, сидящую рядом со мной","Когда я накормил молчаливую девушку, сидящую рядом со мной",隣の席の無口ギャルに餌付けしたら}	'а':1A 'в':12B,20B 'гяру':7A,19B,33B,41B 'девушку':46A,55A 'других':40B 'если':3A,42A 'жизнь':27B 'за':29B 'издевались':18B 'изменить':25B 'кажется':35B 'когда':51A 'мной':11A,16B,50A,59A 'молчаливую':6A,45A,54A 'надо':15B 'накормил':53A 'накормлю':5A,44A 'немного':37B 'но':28B,34B 'оказалась':32B 'она':36B 'от':39B 'отличается':38B 'партой':31B 'постоянно':17B 'решил':24B 'рядом':9A,48A,57A 'свою':26B 'сидящую':8A,47A,56A 'со':10A,49A,58A 'соседней':30B 'средней':13B 'старшей':21B 'что':2A 'школе':14B,22B 'я':4A,23B,43A,52A '隣の席の無口ギャルに餌付けしたら':60A	2026-01-17 22:30:43.304163+00	2026-01-17 22:30:43.304163+00
92a1b7fa-070f-4c56-8e26-afc1db52130c	ru	Обречённый Убийца	«Я вижу число над твоей головой. Я знаю, скольких ты убьешь». Убей первым — и трагедии не произойдет. Следуя этому жестокому правилу, маньяк по имени Номер открывает сезон охоты на «будущих убийц». Он устраняет их до того, как они успеют совершить свое первое преступление. Его цель — идеальный мир без убийств. Его метод — убивать. Сбудется ли мечта того, кто пытается смыть кровь еще большей кровью?	{"Предначертанный убийца",살인예정자}	'без':50B 'большей':64B 'будущих':32B 'вижу':4B 'головой':8B 'до':37B 'его':46B,52B 'еще':63B 'жестокому':22B 'знаю':10B 'и':16B 'идеальный':48B 'имени':26B 'их':36B 'как':39B 'кровь':62B 'кровью':65B 'кто':59B 'ли':56B 'маньяк':24B 'метод':53B 'мечта':57B 'мир':49B 'на':31B 'над':6B 'не':18B 'номер':27B 'обречённый':1A 'он':34B 'они':40B 'открывает':28B 'охоты':30B 'первое':44B 'первым':15B 'по':25B 'правилу':23B 'предначертанный':66A 'преступление':45B 'произойдет':19B 'пытается':60B 'сбудется':55B 'свое':43B 'сезон':29B 'скольких':11B 'следуя':20B 'смыть':61B 'совершить':42B 'твоей':7B 'того':38B,58B 'трагедии':17B 'ты':12B 'убей':14B 'убивать':54B 'убийств':51B 'убийц':33B 'убийца':2A,67A 'убьешь':13B 'успеют':41B 'устраняет':35B 'цель':47B 'число':5B 'этому':21B 'я':3B,9B '살인예정자':68A	2026-01-17 22:30:43.306071+00	2026-01-17 22:30:43.306071+00
e69f0077-8bc6-47a2-9fba-6fa47a926984	ru	Подонок из клана Дан слишком силён	Хилый воин Чу Хонмун втайне овладел секретной техникой рода Дан. Но перед лицом нападения культа Чёрного Пламени оказался бессилен. Скорбя о своём низком происхождении, он погиб.\n\n…Но, открыв глаза, понял: он стал Даном Вольбином — четвёртым сыном рода Дан, печально известным распутником. Он вернулся в прошлое, до гибели семьи, и получил тело с выдающимся даром к боевым искусствам.\n\n«Собственными руками я изменю судьбу и обязательно спасу род Дан!»\n\nВот только в этом чёртовом семействе проблем целая прорва…\n\nТак начинается вторая жизнь отъявленного подонка — со старой памятью, новой силой и железной решимостью всё переписать.	{"The Scoundrel of the Dan Clan Is Too Strong","Данши Сэгье слишком силён","단씨세가 망나니가 너무 강함"}	'clan':104A 'dan':103A 'is':105A 'of':101A 'scoundrel':100A 'strong':107A 'the':99A,102A 'too':106A 'бессилен':25B 'боевым':62B 'в':50B,76B 'вернулся':49B 'воин':8B 'вольбином':40B 'вот':74B 'всё':97B 'втайне':11B 'вторая':85B 'выдающимся':59B 'гибели':53B 'глаза':35B 'дан':4A,16B,44B,73B 'даном':39B 'данши':108A 'даром':60B 'до':52B 'железной':95B 'жизнь':86B 'и':55B,69B,94B 'из':2A 'известным':46B 'изменю':67B 'искусствам':63B 'к':61B 'клана':3A 'культа':21B 'лицом':19B 'нападения':20B 'начинается':84B 'низком':29B 'но':17B,33B 'новой':92B 'о':27B 'обязательно':70B 'овладел':12B 'оказался':24B 'он':31B,37B,48B 'открыв':34B 'отъявленного':87B 'памятью':91B 'перед':18B 'переписать':98B 'печально':45B 'пламени':23B 'погиб':32B 'подонка':88B 'подонок':1A 'получил':56B 'понял':36B 'проблем':80B 'происхождении':30B 'прорва':82B 'прошлое':51B 'распутником':47B 'решимостью':96B 'род':72B 'рода':15B,43B 'руками':65B 'с':58B 'своём':28B 'секретной':13B 'семействе':79B 'семьи':54B 'силой':93B 'силён':6A,111A 'скорбя':26B 'слишком':5A,110A 'со':89B 'собственными':64B 'спасу':71B 'стал':38B 'старой':90B 'судьбу':68B 'сыном':42B 'сэгье':109A 'так':83B 'тело':57B 'техникой':14B 'только':75B 'хилый':7B 'хонмун':10B 'целая':81B 'четвёртым':41B 'чу':9B 'чёрного':22B 'чёртовом':78B 'этом':77B 'я':66B '강함':115A '너무':114A '단씨세가':112A '망나니가':113A	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
d15e5d69-6582-4ac9-97cf-997c49b90aff	ru	Необычные повседневные дела Хатори и Фуруты	Хатори — ничем не примечательный парень, плывущий по течению. Всё меняется, когда он знакомится с Фурутой. Она умеет видеть необычное в самых простых вещах, которые нас окружают. Это (не)обычная история о юности и о том, как маленькие чудеса могут перевернуть целый мир!	{羽鳥と古田の非日常茶飯事}	'в':26B 'вещах':29B 'видеть':24B 'всё':15B 'дела':3A 'знакомится':19B 'и':5A,39B 'история':36B 'как':42B 'когда':17B 'которые':30B 'маленькие':43B 'меняется':16B 'мир':48B 'могут':45B 'нас':31B 'не':9B,34B 'необычное':25B 'необычные':1A 'ничем':8B 'о':37B,40B 'обычная':35B 'окружают':32B 'он':18B 'она':22B 'парень':11B 'перевернуть':46B 'плывущий':12B 'по':13B 'повседневные':2A 'примечательный':10B 'простых':28B 'с':20B 'самых':27B 'течению':14B 'том':41B 'умеет':23B 'фурутой':21B 'фуруты':6A 'хатори':4A,7B 'целый':47B 'чудеса':44B 'это':33B 'юности':38B '羽鳥と古田の非日常茶飯事':49A	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
\.


--
-- Data for Name: novel_proposals; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_proposals (id, user_id, original_link, status, title, alt_titles, author, description, cover_url, genres, tags, vote_score, votes_count, moderator_id, reject_reason, created_at, updated_at, translation_tickets_invested) FROM stdin;
2540f0fd-98d6-4f5b-8514-867bb025d283	ce76495c-3f80-4c7a-be56-3d5994b49b03	https://remanga.org/manga/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/main	draft	Сильнейший убийца и злодейка, убившая Владыку Тьмы, были сосланы на край света	{"Сильнейший убийца и злодейка, убившая Владыку Тьмы"}	zxckotee123	Он - сильнейший убийца, герой войны с армией демона.\nОна - красавица, что собственноручно прикончила Владыку Тьмы и сожгла его замок.\n\nНаграда за спасение мира? Брак по принуждению и ссылка на выжженную окраину. Королевство решило избавиться от слишком опасных героев - тихо и навсегда. Но на руинах замка демона начинается новая история: сюда тянутся люди и демоны, верные лишь силе. А когда король отправляет армию, чтобы «поставить точку», становится ясно:\nошибкой было не наградить их… а попытаться убрать.\n\nИстория о героях, которых выкинули - и которые стали угрозой для всего мира.	https://remanga.org/media/titles/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/cover_08c7cdf2fb8d4a89.webp	{fantasy,romance,thriller}	{academy,regression,monsters,weak_to_strong}	0	0	\N	\N	2026-01-16 20:35:24.250615+00	2026-01-21 21:27:02.819082+00	0
69a35675-51eb-4399-82b3-317245507386	ce76495c-3f80-4c7a-be56-3d5994b49b03	https://remanga.org/manga/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/main	draft	Сильнейший убийца и злодейка, убившая Владыку Тьмы, были сосланы на край света	{}	zxc111	Хз... Ну что, погнали господа!\nЗавезите фуру лайков и оценок, да будет вам продолжение. Мы догнали онгоинг, следующая глава выйдет 18 числа 	https://remanga.org/media/titles/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/cover_08c7cdf2fb8d4a89.webp	{action,scifi,martial_arts}	{transmigration,female_lead,mature}	0	0	\N	\N	2026-01-17 23:47:15.033486+00	2026-01-21 21:27:02.819082+00	0
d19d3e8f-a609-41ad-b5db-ca0f9ca09b24	ce76495c-3f80-4c7a-be56-3d5994b49b03	https://remanga.org/manga/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/main	draft	Сильнейший убийца и злодейка, убившая Владыку Тьмы, были сосланы на край света	{}	zxc111	Хз... Ну что, погнали господа!\nЗавезите фуру лайков и оценок, да будет вам продолжение. Мы догнали онгоинг, следующая глава выйдет 18 числа 	https://remanga.org/media/titles/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/cover_08c7cdf2fb8d4a89.webp	{action,scifi,martial_arts}	{transmigration,female_lead,mature}	0	0	\N	\N	2026-01-17 23:47:24.201821+00	2026-01-21 21:27:02.819082+00	0
2074314d-bf4a-4c22-b8cd-d07d086ca73a	ce76495c-3f80-4c7a-be56-3d5994b49b03	https://remanga.org/manga/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/main	draft	Сильнейший убийца и злодейка, убившая Владыку Тьмы, были сосланы на край света	{}	zxc111	Хз... Ну что, погнали господа!\nЗавезите фуру лайков и оценок, да будет вам продолжение. Мы догнали онгоинг, следующая глава выйдет 18 числа 	https://remanga.org/media/titles/the-strongest-assassin-uncle-after-saving-the-world-was-pressed-by-the-villain-daughter-of-the-demon-king-killing-became-a-frontier-lord/cover_08c7cdf2fb8d4a89.webp	{action,scifi,martial_arts}	{transmigration,female_lead,mature}	0	0	\N	\N	2026-01-17 23:49:21.476567+00	2026-01-21 21:27:02.819082+00	0
fd3b9c09-7f81-41c4-a459-3a2db58420f4	ce76495c-3f80-4c7a-be56-3d5994b49b03	https://remanga.org/manga/the-strongest-dull-princes-secret-battle-for-the-throne/main	voting	Тайная Битва за Престол Сильнейшего Принца-Авантюриста SS-ранга	{"The Strongest Dull Prince’s Secret Battle for the Throne  Saikyou Degarashi Ouji no An'yaku Teii Arasoi, Saikyou Degarashi Ouji no An'yaku Teii Arasoi Munou wo Enjiru SS Rank Ouji wa Koui Keishou-sen wo Kage kara Shihai Suru  Saikyou Degarashi Ouji no An'yaku Teii Arasoi ~Munou wo Enjiru SS Rank Ouji wa Koui Keishou-sen wo Kage kara Shihai Suru~  最強出涸らし皇子の暗躍帝位争い"}	ячс123	Империя Эдразия на контитенте Вогел. С обширными землями, и огромной военной силой, она раздирается в борьбе за престол.Пока наследник не определен, дети императора изо всех сил стараются отхватить как можно больше власти. Однако, есть среди них один принц, который утверждает, что ни при каких обстоятельствах не хочет быть императором.Седьмой принц, Арнольд Лэйкс Адлер. Юноша, который во всем уступает своему младшему брату-близнецу, принц-дуралей.Апатичный и бездарный, Арнольд все время тратит, слоняясь без дела. Однако, за ширмой он один из всего пяти авантюристов SS класса, именуемый Сильвером.Глядя на накаляющуюся борьбу за престол, он решил: "Не хочу помирать, так что сделаю лучше императором младшего братца...."И началась история абсурдных тайных интриг принца, которому было плевать на императорскй трон.	https://remanga.org/media/titles/the-strongest-dull-princes-secret-battle-for-the-throne/cover_6c1f6430.webp	{xuanhuan,scifi,action}	{sword_and_sorcery,magic,strong_mc}	2	2	ce76495c-3f80-4c7a-be56-3d5994b49b03	\N	2026-01-17 23:56:22.040192+00	2026-01-22 20:37:09.171575+00	6
\.


--
-- Data for Name: novel_ratings; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_ratings (novel_id, user_id, value, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: novel_tags; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_tags (novel_id, tag_id) FROM stdin;
\.


--
-- Data for Name: novel_views_daily; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novel_views_daily (novel_id, date, views) FROM stdin;
\.


--
-- Data for Name: novels; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.novels (id, slug, cover_image_key, translation_status, original_chapters_count, release_year, author, views_total, views_daily, rating_sum, rating_count, bookmarks_count, created_at, updated_at) FROM stdin;
6db455a3-5bf9-4e65-a07a-0f6f3118a452	mangalib-254686-254686-gyaru-shimai-wa-botchina-ne-sumi-kun-o-kamaitai	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.080783+00	2026-01-17 22:24:28.080783+00
5be035c3-ec48-445d-88e1-b08c1b6f1bb2	mangalib-242614-242614-dungeon-crawler-carl	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.089772+00	2026-01-17 22:24:28.089772+00
b5818047-3d8a-4700-af9a-1758c1db91c5	mangalib-253852-253852-one-high-school-hero	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.091236+00	2026-01-17 22:24:28.091236+00
bd31a32c-1889-4016-80df-631d232bdc7f	mangalib-244624-244624-sikeulis-paemilli	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.092656+00	2026-01-17 22:24:28.092656+00
6d8c6389-1ca6-4c9f-8212-1c94b880546d	mangalib-220281-220281-isanghande-hyogwajeog-in-agnyeo-saenghwal	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.09436+00	2026-01-17 22:24:28.09436+00
da4defbb-aef7-4b9b-a1a4-ec6f96172bc0	mangalib-223027-223027-naui-jjabsalang-sunjeong	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.095973+00	2026-01-17 22:24:28.095973+00
f0b74a56-9c98-4354-87db-45e7808e271e	mangalib-251286-251286-naui-yeon-ae-chamgyeonja	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.097565+00	2026-01-17 22:24:28.097565+00
7c28c91b-ebdd-4407-b3e9-c74e7c9ff980	mangalib-196019-196019-warriors-the-prophecies-begin	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.098915+00	2026-01-17 22:24:28.098915+00
0097c045-1941-4706-adb9-b1b592200c7e	mangalib-237160-237160-the-fake-master-who-accidentally-became-the-strongest	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.100432+00	2026-01-17 22:24:28.100432+00
1b3a53fc-59ff-46ae-ab60-4b9f50e3ae19	mangalib-219704-219704-yug-amul-eommaneun-kkog-jug-eoya-hanayo	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.101758+00	2026-01-17 22:24:28.101758+00
a4b9d9d2-1ac0-40e7-9a22-57d5c2addf6f	mangalib-222507-222507-moleolliseu-seukaendeul	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.103183+00	2026-01-17 22:24:28.103183+00
c66807fe-8d42-4290-8bbc-f0a8bf807c3a	mangalib-243149-243149-i-wanna-eat-your-guts	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.104482+00	2026-01-17 22:24:28.104482+00
8e9c621f-c81d-43ca-95a5-b662cdb79516	mangalib-242622-242622-bing-uihan-agnyeoga-kkaengpan-chim	\N	ongoing	0	\N	\N	0	0	0	0	0	2026-01-17 22:24:28.105791+00	2026-01-17 22:24:28.105791+00
72cd95ba-f0f0-4330-98a4-f171c7f8cabe	mangalib-181356-181356-byeongyeongbaeg-seojaneun-hwangjeyeossda	covers/72cd95ba-f0f0-4330-98a4-f171c7f8cabe.jpg	ongoing	96	\N	\N	0	0	0	0	0	2026-01-17 22:30:43.26705+00	2026-01-17 22:30:43.26705+00
a239b948-358f-4179-808d-f8c6bb5715c3	mangalib-239594-239594-tonari-no-seki-no-mukuchi-gyaru-ni-edzuke-shitara	covers/a239b948-358f-4179-808d-f8c6bb5715c3.jpg	ongoing	1	\N	\N	0	0	0	0	0	2026-01-17 22:30:43.304163+00	2026-01-17 22:30:43.304163+00
e69f0077-8bc6-47a2-9fba-6fa47a926984	mangalib-244682-244682-danssi-sega-mangnani-ga-neomu-gangham	covers/e69f0077-8bc6-47a2-9fba-6fa47a926984.jpg	ongoing	13	\N	\N	0	0	0	0	0	2026-01-17 22:30:43.310948+00	2026-01-17 22:30:43.310948+00
d15e5d69-6582-4ac9-97cf-997c49b90aff	mangalib-252235-252235-hatori-to-furuta-no-hi-nichijochahanji	covers/d15e5d69-6582-4ac9-97cf-997c49b90aff.jpg	ongoing	4	\N	\N	0	0	0	0	0	2026-01-17 22:30:43.315092+00	2026-01-17 22:30:43.315092+00
0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	mangalib-253241-253241-moriagaranai-deto	covers/0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00.jpg	ongoing	1	\N	\N	15	15	0	0	0	2026-01-17 22:30:43.299741+00	2026-01-22 20:48:20.214915+00
92a1b7fa-070f-4c56-8e26-afc1db52130c	mangalib-223436-223436-sal-in-yejeongja	covers/92a1b7fa-070f-4c56-8e26-afc1db52130c.jpg	ongoing	16	\N	\N	3	3	0	0	0	2026-01-17 22:30:43.306071+00	2026-01-22 21:00:51.154505+00
6dad104f-6d7e-4c26-927e-e4457ab8fb1b	mangalib-249628-249628-aikaeschiyu	covers/6dad104f-6d7e-4c26-927e-e4457ab8fb1b.jpg	ongoing	4	\N	\N	9	9	0	0	0	2026-01-17 22:30:43.301742+00	2026-01-22 21:16:47.703292+00
\.


--
-- Data for Name: platform_stats; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.platform_stats (id, total_novels, total_chapters, total_users, total_comments, total_collections, total_votes_cast, total_tickets_spent, proposals_translated, updated_at) FROM stdin;
1	0	0	0	0	0	0	0	0	2026-01-11 16:04:08.442605+00
\.


--
-- Data for Name: reading_progress; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.reading_progress (user_id, novel_id, chapter_id, "position", updated_at) FROM stdin;
ce76495c-3f80-4c7a-be56-3d5994b49b03	0d7a30e9-6fec-4dc4-a6bb-7f0d06689e00	3debf8c8-e80f-42d2-9bc4-2d8950de8e06	0	2026-01-22 20:34:54.675461+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	92a1b7fa-070f-4c56-8e26-afc1db52130c	725c4251-57b6-487e-a5ac-e69c5b9c5ce2	0	2026-01-22 21:01:00.846211+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	6dad104f-6d7e-4c26-927e-e4457ab8fb1b	d2659d23-21ab-4664-a710-a5b54fabe171	0	2026-01-22 21:17:01.206139+00
\.


--
-- Data for Name: refresh_tokens; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked_at) FROM stdin;
66e78478-d055-4fad-be65-8f213a92ba6f	ce76495c-3f80-4c7a-be56-3d5994b49b03	4f51471d3acb294f105c664f5c5f4c8ba4db4e024791d63d9e29bd92da19a115	2026-01-18 16:25:35.377626+00	2026-01-11 16:25:35.377626+00	2026-01-11 16:36:11.2145+00
d86770d9-40f4-48f5-8ad0-641f521dabaa	ce76495c-3f80-4c7a-be56-3d5994b49b03	45b52e05014adff743b26b61d954bf8c2f2b72c1a5667da62efbad0e794f3214	2026-01-18 16:36:11.215659+00	2026-01-11 16:36:11.215659+00	2026-01-11 16:37:03.559489+00
d4597235-869f-4174-8da5-48390ded3325	ce76495c-3f80-4c7a-be56-3d5994b49b03	91e190dce45ee8c8edda2b7b82e3c79a1bc69c261f56ae60099f6b844d6ef4b7	2026-01-18 16:37:03.56058+00	2026-01-11 16:37:03.56058+00	2026-01-11 16:37:45.92641+00
8601ad3b-d5ca-471b-b26f-ffccf7a74c7f	ce76495c-3f80-4c7a-be56-3d5994b49b03	4672591463c6f639e50d100850eb87b1f8adee1cbd3a5ac3fde199b8cd11cafc	2026-01-18 16:37:45.934175+00	2026-01-11 16:37:45.934175+00	2026-01-11 16:41:31.535327+00
10d0eb72-be2e-4252-a0a0-2732183699ba	ce76495c-3f80-4c7a-be56-3d5994b49b03	a33fc2564fd12529b1c9552c20d8432f19b0ec2158e75de65fa687d42931a0bb	2026-01-18 16:41:31.544138+00	2026-01-11 16:41:31.544138+00	\N
33c8977b-1242-4b36-8525-ca9f7be7a818	ce76495c-3f80-4c7a-be56-3d5994b49b03	eb3efab289330372850b705a97f6222eaa58b5d58a95c421bc7a8776777a54ec	2026-01-18 16:41:41.098666+00	2026-01-11 16:41:41.098666+00	2026-01-11 16:41:41.193843+00
b2450580-0301-4a51-b7c5-11e7f71d9244	ce76495c-3f80-4c7a-be56-3d5994b49b03	c3b88dde666f5e5ed48c08720c25aa15290f21fbe55a0d58bfbaad96e59940eb	2026-01-18 16:41:42.198636+00	2026-01-11 16:41:42.198636+00	2026-01-11 16:41:42.284303+00
67b371c6-a655-4c6b-8292-97326a41b2b1	ce76495c-3f80-4c7a-be56-3d5994b49b03	d2006b8ec088af235853cc59e53773004d24641d4fef966b3d34afb3080afe45	2026-01-18 16:41:43.222067+00	2026-01-11 16:41:43.222067+00	2026-01-11 16:41:43.319408+00
c2f67d68-4ccd-4d65-aa3d-7c09d295086e	ce76495c-3f80-4c7a-be56-3d5994b49b03	524b51426954eef0bf5b96bb8ff4a6b5fdbaac1857250fdc4ed7954f1930deb1	2026-01-18 16:43:45.832103+00	2026-01-11 16:43:45.832103+00	2026-01-11 16:43:45.943812+00
a0eaa754-7d70-40c8-b782-afd568f071fe	ce76495c-3f80-4c7a-be56-3d5994b49b03	7b45a227f980f3f53dd945fac4ccf8ec65c0f3e8a9a18b84bd617568af907a8b	2026-01-18 16:43:47.258011+00	2026-01-11 16:43:47.258011+00	2026-01-11 16:43:47.344537+00
ea3a8824-e052-4507-a227-53421bc375e9	ce76495c-3f80-4c7a-be56-3d5994b49b03	0a5105ed4e3ea5cb83312ade990b1097c06bfc48b8395ac80d51adf9548d9f3e	2026-01-18 16:49:11.009082+00	2026-01-11 16:49:11.009082+00	2026-01-11 16:49:14.656149+00
3faf366c-0af7-4acb-8653-2b408db3e6a8	ce76495c-3f80-4c7a-be56-3d5994b49b03	cc8b78b6b83ab1ed05c0d0cc0b9098fd7ae72bcfcbf174f392fc15fd205b43c6	2026-01-18 16:49:14.657566+00	2026-01-11 16:49:14.657566+00	\N
24e34efb-1ee3-4c10-9bc6-4f69f8acf2c8	ce76495c-3f80-4c7a-be56-3d5994b49b03	ce77f6a75234bf7f5131926ea32209cb8c8c009fed6853fdebc17b918f5cf160	2026-01-18 16:49:20.293711+00	2026-01-11 16:49:20.293711+00	2026-01-11 16:49:20.389329+00
785f7191-b562-4007-a84c-91f11c86d419	ce76495c-3f80-4c7a-be56-3d5994b49b03	7ca4b3ee032136a31d763b44d152497b590dbdbdc7d87bc4c68bbd2e6736d094	2026-01-18 16:49:24.666733+00	2026-01-11 16:49:24.666733+00	2026-01-11 16:49:24.744621+00
ee900e04-9d72-49ae-aa03-10528eae4dda	ce76495c-3f80-4c7a-be56-3d5994b49b03	b560ca3553ae2f20956b1294f03276c550a7f7e158ec068d6825eaac1674c7f1	2026-01-18 16:51:42.692419+00	2026-01-11 16:51:42.692419+00	2026-01-11 16:51:42.771852+00
789f1355-937a-4dda-bdaa-f6f4ae3895f5	ce76495c-3f80-4c7a-be56-3d5994b49b03	461b3ef1952be926fc6c59315d07aafbc3e5a3d3e582e2cd12a2cdccebd1b57f	2026-01-18 16:51:43.958567+00	2026-01-11 16:51:43.958567+00	2026-01-11 16:51:44.039426+00
a1d6778f-b0e5-48e7-a702-6b18e7e78e1c	ce76495c-3f80-4c7a-be56-3d5994b49b03	4bc0611388f72ad7ff0d2db7b5140e5ff2c78188d1eec1aef0a024bf8628320d	2026-01-18 16:51:44.040669+00	2026-01-11 16:51:44.040669+00	\N
6866fda1-7754-4aed-b1d8-9dc90b0ee418	ce76495c-3f80-4c7a-be56-3d5994b49b03	e710b5ee69d63216f566dab9a1ac0357f02aa55c85f6e221623403fd016308fb	2026-01-18 16:51:47.87941+00	2026-01-11 16:51:47.87941+00	2026-01-11 20:23:05.840881+00
dfea8bb3-25bc-4ad1-88a0-c65e46cb6277	ce76495c-3f80-4c7a-be56-3d5994b49b03	4f81bdd57c7c52de68aaf8b89358b15a5b781196beadef2cae277f067f5dc4c9	2026-01-18 20:23:05.842021+00	2026-01-11 20:23:05.842021+00	2026-01-11 21:25:48.282084+00
2762a1ff-41b6-43e0-933f-a78b0fc06d44	ce76495c-3f80-4c7a-be56-3d5994b49b03	0ae6ccdff4370b13a1670cb5eccd1f5da80ce4ec169f6c8c98ad73ddcb07804c	2026-01-18 21:25:48.283265+00	2026-01-11 21:25:48.283265+00	\N
ab426154-0360-4800-89d8-7b651a46068a	ce76495c-3f80-4c7a-be56-3d5994b49b03	b8ade61ff2a78ca9f6534ab42334219ceb92ad49ccbd26a1c5007432a90bd763	2026-01-18 21:43:53.680795+00	2026-01-11 21:43:53.680795+00	2026-01-16 19:31:24.908266+00
bb7aa48b-5f49-46bc-8321-1714f0e0693a	ce76495c-3f80-4c7a-be56-3d5994b49b03	9dc395967b962729eea9471fb533c15c0b8f3d731bf7310fc4292bc520a9ae55	2026-01-23 19:31:24.910187+00	2026-01-16 19:31:24.910187+00	\N
d00dd614-f515-4de6-a1db-d1b13ab92970	ce76495c-3f80-4c7a-be56-3d5994b49b03	2076b21c7f3a4d871435991a66684175888eb316195bcd031ec10442a550bd91	2026-01-23 20:02:43.523247+00	2026-01-16 20:02:43.523247+00	\N
7691871e-bc54-499f-9ea4-fee94523a4dd	30fa06a3-de30-45ee-a148-f222c1da6e20	9016ce316b592477ba0aa8dd0a2b0f9a73992b84868d3a267619aca0831a4e9a	2026-01-23 20:36:56.694914+00	2026-01-16 20:36:56.694914+00	\N
316a631b-8e14-4e1c-9bb7-eacfcaedee83	30fa06a3-de30-45ee-a148-f222c1da6e20	3d843f2ce70f0c090bf706d386e69124f5e7a6e4d83fadf99754eebbaef39d2c	2026-01-23 20:38:36.684326+00	2026-01-16 20:38:36.684326+00	\N
db639799-59b5-414e-b6a0-7529e2b71027	ce76495c-3f80-4c7a-be56-3d5994b49b03	85a162cd855c36d3e08f445eab8c7b2feeba95cb5387680cf02652d632ce5fa7	2026-01-23 21:02:12.446709+00	2026-01-16 21:02:12.446709+00	\N
4d498216-faa8-42d6-9fc0-ddc699b88bc1	ce76495c-3f80-4c7a-be56-3d5994b49b03	a327ba8a27536713a07fbacb7aff822f06f0265cfc767c2abdd46c8fa18a8483	2026-01-23 21:15:15.148198+00	2026-01-16 21:15:15.148198+00	\N
e6d15be2-e8fc-4027-81e7-ae2a6fc0ce5f	ce76495c-3f80-4c7a-be56-3d5994b49b03	6d73b6e1fd47159ba416e41cb2a45fd3b80e4aec47a4c4948c60cac072317ec6	2026-01-23 22:42:14.436752+00	2026-01-16 22:42:14.436752+00	2026-01-16 23:49:00.835245+00
8074bb77-ca89-4fb1-ae0c-cec8f2d09db8	ce76495c-3f80-4c7a-be56-3d5994b49b03	fc75d397c7de5f0a3c4bc0b7a0e6a2dc4102de6c76b2558588eeb5e293433157	2026-01-23 23:49:00.836338+00	2026-01-16 23:49:00.836338+00	2026-01-17 23:35:35.34032+00
3069b5af-e8ea-48b3-a3b3-307ebd332b58	ce76495c-3f80-4c7a-be56-3d5994b49b03	9b88d68970766822b03613bcf6d510a1c03dbe8d418bc07d47c1cf7e319013cc	2026-01-24 23:35:35.341337+00	2026-01-17 23:35:35.341337+00	2026-01-21 19:54:21.933691+00
fe47c62e-b248-4bc4-8f0c-c0bf4c655d4e	ce76495c-3f80-4c7a-be56-3d5994b49b03	9c6682c8e241618ecddc16edea4d609517bffc13654920d3594ce1f5132467ac	2026-01-28 19:54:21.935437+00	2026-01-21 19:54:21.935437+00	\N
1fa0be88-da17-482b-977d-802e9d4766f0	30fa06a3-de30-45ee-a148-f222c1da6e20	cc37eff3ada1d93ea1362187c359b8ea0c0cd14d53c4706a7cd4be94a2ad03f8	2026-01-28 20:39:24.769336+00	2026-01-21 20:39:24.769336+00	\N
90c665f7-2d51-46be-be8b-0da0d5e32f51	30fa06a3-de30-45ee-a148-f222c1da6e20	41d846eddb3ea17566e40635e45c143f28e6032484f1e20956143216fab64547	2026-01-28 20:39:42.15613+00	2026-01-21 20:39:42.15613+00	\N
ead29fba-19e8-4001-bee3-5f490a5e45b7	ce76495c-3f80-4c7a-be56-3d5994b49b03	a595ddbc6e3f3c31ad89761737f3438c761115b2bcc32feaea43a0d3e90b01ab	2026-01-28 20:39:59.413444+00	2026-01-21 20:39:59.413444+00	\N
b4c60cc5-52fd-4b03-922b-982d3ec95169	ce76495c-3f80-4c7a-be56-3d5994b49b03	ae000ee2fadc6e6d0c92720eaab19c83c3ab246dc4dc81b0c354976d2294778b	2026-01-28 20:55:08.763863+00	2026-01-21 20:55:08.763863+00	\N
559538d8-ccfa-4632-bee0-a44443c5927a	30fa06a3-de30-45ee-a148-f222c1da6e20	c899ffd2021f4b28019d7a7ca170cf0095684cddd186ca75267cab1b834cd895	2026-01-28 20:56:09.177237+00	2026-01-21 20:56:09.177237+00	\N
a1679fb1-b6b5-4a21-b03c-4e38496f11dd	ce76495c-3f80-4c7a-be56-3d5994b49b03	b2fd27534a695fd4d3fef5fffb1f376112d42a9693b482a19fac4212356c68eb	2026-01-28 21:30:05.539166+00	2026-01-21 21:30:05.539166+00	2026-01-22 20:20:25.219873+00
95ce39be-c51f-4d1a-a998-cd025bbc68d7	ce76495c-3f80-4c7a-be56-3d5994b49b03	83553b393916a9b28a060988ef7b83824e139bcb22f58b3fe1104645458018f1	2026-01-29 20:20:25.221042+00	2026-01-22 20:20:25.221042+00	\N
6c264dd9-2359-4949-9f12-38e52c6b78a2	30fa06a3-de30-45ee-a148-f222c1da6e20	c22f4d1d5b45001dc200fd242bcbb8d8dad8a9646b05a091bb673aeb8a983fed	2026-01-29 20:36:59.11272+00	2026-01-22 20:36:59.11272+00	\N
4f54de71-d8ee-45a3-a657-3998620b6ee7	ce76495c-3f80-4c7a-be56-3d5994b49b03	0a3b47bff11622d3fb100f9a6143f625f47bbd6d6a3346cd5a155d917eeb4cf2	2026-01-29 20:41:53.743636+00	2026-01-22 20:41:53.743636+00	\N
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
001_initial_schema	2026-01-11 15:45:47.309833+00
002_mvp2_comments_bookmarks_xp	2026-01-11 16:04:08.191329+00
003_economy	2026-01-11 16:04:08.297379+00
004_community	2026-01-11 16:04:08.442605+00
005_remove_bookmark_list_title	2026-01-11 21:26:31.331836+00
006_authors	2026-01-16 22:42:07.835258+00
007_comments_unify	2026-01-16 22:42:07.877412+00
008_admin_settings_and_audit	2026-01-16 22:42:07.900141+00
009_weekly_ticket_grants_and_plans	2026-01-21 21:08:10.854156+00
010_translation_ticket_investments	2026-01-21 21:27:02.819082+00
011_comment_anchor	2026-01-21 21:49:24.427527+00
\.


--
-- Data for Name: subscription_grants; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.subscription_grants (id, subscription_id, user_id, type, amount, for_month, granted_at) FROM stdin;
7576ae2a-19d1-4214-95eb-837c51243d35	36aede9e-786f-4559-bcbc-facc8c6dfcb5	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	3	2026-01	2026-01-11 20:51:41.931571+00
2808d274-a64a-4998-a38e-11ee84a19ea0	36aede9e-786f-4559-bcbc-facc8c6dfcb5	ce76495c-3f80-4c7a-be56-3d5994b49b03	translation_ticket	10	2026-01	2026-01-11 20:51:41.934331+00
4dc48fc5-8fec-4752-a395-871ee3c14bd9	3121636e-672a-4c3e-8c92-887532999384	30fa06a3-de30-45ee-a148-f222c1da6e20	novel_request	3	2026-01	2026-01-16 20:45:56.940464+00
487a5882-51f6-45ca-bd96-d71d32d9f329	3121636e-672a-4c3e-8c92-887532999384	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	10	2026-01	2026-01-16 20:45:56.943095+00
\.


--
-- Data for Name: subscription_plans; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.subscription_plans (id, code, title, description, price, currency, period, is_active, features, created_at, updated_at) FROM stdin;
c5851ba5-be78-48f7-a816-9ff8d7850fb1	basic	Basic	Базовая подписка с отключением рекламы	19900	RUB	monthly	t	{"adFree": true, "exclusiveBadge": false, "prioritySupport": false, "canEditDescriptions": false, "dailyVoteMultiplier": 1, "monthlyNovelRequests": 1, "canRequestRetranslation": false, "monthlyTranslationTickets": 3}	2026-01-11 16:04:08.297379+00	2026-01-11 16:04:08.297379+00
74a8c0a0-95e2-4545-b5c6-04753d59af1c	premium	Premium	Премиум подписка с расширенными возможностями	49900	RUB	monthly	t	{"adFree": true, "exclusiveBadge": true, "prioritySupport": false, "canEditDescriptions": true, "dailyVoteMultiplier": 2, "monthlyNovelRequests": 2, "canRequestRetranslation": true, "monthlyTranslationTickets": 5}	2026-01-11 16:04:08.297379+00	2026-01-21 21:08:10.854156+00
fb090153-96e9-4837-8e37-314aaf936089	vip	VIP	Максимальная подписка со всеми привилегиями	99900	RUB	monthly	t	{"adFree": true, "exclusiveBadge": true, "prioritySupport": true, "canEditDescriptions": true, "dailyVoteMultiplier": 5, "monthlyNovelRequests": 5, "canRequestRetranslation": true, "monthlyTranslationTickets": 15}	2026-01-11 16:04:08.297379+00	2026-01-21 21:08:10.854156+00
\.


--
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.subscriptions (id, user_id, plan_id, status, starts_at, ends_at, external_id, auto_renew, canceled_at, created_at, updated_at) FROM stdin;
36aede9e-786f-4559-bcbc-facc8c6dfcb5	ce76495c-3f80-4c7a-be56-3d5994b49b03	74a8c0a0-95e2-4545-b5c6-04753d59af1c	active	2026-01-11 20:51:41.918064+00	2026-02-11 20:51:41.918064+00	\N	f	\N	2026-01-11 20:51:41.918066+00	2026-01-11 20:51:41.918066+00
3121636e-672a-4c3e-8c92-887532999384	30fa06a3-de30-45ee-a148-f222c1da6e20	74a8c0a0-95e2-4545-b5c6-04753d59af1c	active	2026-01-16 20:45:56.928656+00	2026-02-16 20:45:56.928656+00	\N	f	\N	2026-01-16 20:45:56.92866+00	2026-01-16 20:45:56.92866+00
\.


--
-- Data for Name: tag_localizations; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.tag_localizations (tag_id, lang, name) FROM stdin;
\.


--
-- Data for Name: tags; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.tags (id, slug, created_at) FROM stdin;
\.


--
-- Data for Name: ticket_balances; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.ticket_balances (user_id, type, balance, updated_at) FROM stdin;
ce76495c-3f80-4c7a-be56-3d5994b49b03	daily_vote	2	2026-01-21 20:37:34.591171+00
30fa06a3-de30-45ee-a148-f222c1da6e20	novel_request	5	2026-01-21 21:28:19.105966+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	2	2026-01-21 21:28:19.109225+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	translation_ticket	15	2026-01-21 21:28:19.110258+00
30fa06a3-de30-45ee-a148-f222c1da6e20	daily_vote	0	2026-01-22 20:37:03.677166+00
30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	9	2026-01-22 20:37:09.179368+00
\.


--
-- Data for Name: ticket_transactions; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.ticket_transactions (id, user_id, type, delta, reason, ref_type, ref_id, idempotency_key, created_at) FROM stdin;
cb4cf716-b351-4e26-9194-ae0522fd4f81	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	3	subscription_grant	subscription	36aede9e-786f-4559-bcbc-facc8c6dfcb5	sub_grant:36aede9e-786f-4559-bcbc-facc8c6dfcb5:2026-01:novel_request	2026-01-11 20:51:41.928087+00
e5e35211-2485-4906-aef2-2d2d21c674b6	ce76495c-3f80-4c7a-be56-3d5994b49b03	translation_ticket	10	subscription_grant	subscription	36aede9e-786f-4559-bcbc-facc8c6dfcb5	sub_grant:36aede9e-786f-4559-bcbc-facc8c6dfcb5:2026-01:translation	2026-01-11 20:51:41.933231+00
b249c454-0192-4495-877a-ce2a6ca0d2c3	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	-1	proposal_created	proposal	2540f0fd-98d6-4f5b-8514-867bb025d283		2026-01-16 20:35:24.263445+00
bc638927-324f-445b-88e4-30819e88e406	30fa06a3-de30-45ee-a148-f222c1da6e20	novel_request	3	subscription_grant	subscription	3121636e-672a-4c3e-8c92-887532999384	sub_grant:3121636e-672a-4c3e-8c92-887532999384:2026-01:novel_request	2026-01-16 20:45:56.938131+00
12a217b6-ea18-423e-aa87-3e5a8c80a0f8	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	10	subscription_grant	subscription	3121636e-672a-4c3e-8c92-887532999384	sub_grant:3121636e-672a-4c3e-8c92-887532999384:2026-01:translation	2026-01-16 20:45:56.941969+00
1ff5802d-3094-4432-b01d-b57d0c9a748a	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	-1	proposal_created	proposal	2074314d-bf4a-4c22-b8cd-d07d086ca73a	\N	2026-01-17 23:49:21.489351+00
e14a9e3d-cfc6-4324-ae9e-69f09be3a789	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	-1	proposal_created	proposal	fd3b9c09-7f81-41c4-a459-3a2db58420f4	\N	2026-01-17 23:56:22.049263+00
5cf4dc0e-ec2e-4387-8844-b600be38236f	30fa06a3-de30-45ee-a148-f222c1da6e20	daily_vote	2	daily_grant	daily_grant	\N	daily_vote:2026-01-21:30fa06a3-de30-45ee-a148-f222c1da6e20	2026-01-21 20:37:34.588086+00
7c79ed6d-272d-4d3d-9207-84e4592595c0	ce76495c-3f80-4c7a-be56-3d5994b49b03	daily_vote	2	daily_grant	daily_grant	\N	daily_vote:2026-01-21:ce76495c-3f80-4c7a-be56-3d5994b49b03	2026-01-21 20:37:34.591293+00
0cb264c4-29c6-4488-abcd-c57ae956dce0	30fa06a3-de30-45ee-a148-f222c1da6e20	daily_vote	-1	vote_cast	vote	8eec8435-7846-4e5f-9261-42ea706cfa82	\N	2026-01-21 20:56:16.485235+00
c956fdc9-2c0d-454b-a8ef-fcd615a95bb3	30fa06a3-de30-45ee-a148-f222c1da6e20	novel_request	2	subscription_grant	weekly_grant	\N	weekly_sub:2026-01-21:30fa06a3-de30-45ee-a148-f222c1da6e20:novel_request	2026-01-21 21:28:19.105867+00
5dd8dbe6-fd59-4620-8c81-73be5e85deef	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	5	subscription_grant	weekly_grant	\N	weekly_sub:2026-01-21:30fa06a3-de30-45ee-a148-f222c1da6e20:translation_ticket	2026-01-21 21:28:19.108042+00
9342e5da-ba87-4950-ae9f-70ed7cea21f0	ce76495c-3f80-4c7a-be56-3d5994b49b03	novel_request	2	subscription_grant	weekly_grant	\N	weekly_sub:2026-01-21:ce76495c-3f80-4c7a-be56-3d5994b49b03:novel_request	2026-01-21 21:28:19.109108+00
4cda39d1-0cc3-4799-adad-11b8a7d02f11	ce76495c-3f80-4c7a-be56-3d5994b49b03	translation_ticket	5	subscription_grant	weekly_grant	\N	weekly_sub:2026-01-21:ce76495c-3f80-4c7a-be56-3d5994b49b03:translation_ticket	2026-01-21 21:28:19.110184+00
98c2970f-43c0-4058-84af-869432e9db58	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	e3de0376-c2f2-4ab2-88f4-df1c45aff7ca	\N	2026-01-21 21:29:24.297831+00
6ba0b2cf-420c-4390-bd08-b9159f47e2ba	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	b8be3e4d-26b3-492b-9e76-e0a05499a0af	\N	2026-01-21 21:29:24.975096+00
8a3be05e-f4e5-4c4c-9c72-ee0545617f56	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	863179f1-948a-4362-99a0-52d935a8d0dd	\N	2026-01-21 21:29:25.1582+00
49843058-3214-4651-8b57-696539cf8181	30fa06a3-de30-45ee-a148-f222c1da6e20	daily_vote	-1	vote_cast	vote	9e6da1a4-ce8c-4044-8630-7b7a5cd62b4d	\N	2026-01-22 20:37:03.677533+00
5091c88d-b038-499b-aaf1-5ac5ca3cb67a	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	41811167-e837-46f9-95e5-eab6904932a5	\N	2026-01-22 20:37:08.47961+00
eee94c45-b7c1-4f1e-a3b2-93fa72a6066a	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	b2085ac0-be13-452a-836b-ec17e2e30b6f	\N	2026-01-22 20:37:08.693704+00
438e9351-9575-4f80-ae5f-433c1cf7ead7	30fa06a3-de30-45ee-a148-f222c1da6e20	translation_ticket	-1	vote_cast	vote	5e30b6e9-b296-4317-ab71-472c1ae98f79	\N	2026-01-22 20:37:09.179613+00
\.


--
-- Data for Name: user_achievements; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.user_achievements (user_id, achievement_id, unlocked_at) FROM stdin;
\.


--
-- Data for Name: user_profiles; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.user_profiles (user_id, display_name, avatar_key, bio, created_at, updated_at) FROM stdin;
ce76495c-3f80-4c7a-be56-3d5994b49b03	zxckotee	\N	\N	2026-01-11 16:25:35.373207+00	2026-01-11 16:25:35.373207+00
30fa06a3-de30-45ee-a148-f222c1da6e20	zxckotee2	\N	\N	2026-01-16 20:36:56.689847+00	2026-01-16 20:36:56.689847+00
\.


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.user_roles (user_id, role, created_at) FROM stdin;
ce76495c-3f80-4c7a-be56-3d5994b49b03	premium	2026-01-11 20:28:13.578146+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	admin	2026-01-16 19:33:47.29366+00
30fa06a3-de30-45ee-a148-f222c1da6e20	user	2026-01-16 20:36:56.689926+00
30fa06a3-de30-45ee-a148-f222c1da6e20	premium	2026-01-16 20:38:24.709174+00
\.


--
-- Data for Name: user_xp; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.user_xp (user_id, xp_total, level, updated_at) FROM stdin;
ce76495c-3f80-4c7a-be56-3d5994b49b03	60	1	2026-01-22 20:35:00.838051+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.users (id, email, password_hash, is_banned, last_login_at, created_at, updated_at) FROM stdin;
30fa06a3-de30-45ee-a148-f222c1da6e20	zxckotee2@yandex.ru	$2a$10$r6b.VIjpMLsowysaoag9OundjVcqsOhnIoe3z6yc3rT4KYdjHAhAS	f	2026-01-22 20:36:59.110373+00	2026-01-16 20:36:56.689847+00	2026-01-22 20:36:59.110531+00
ce76495c-3f80-4c7a-be56-3d5994b49b03	zxckotee@yandex.ru	$2a$10$B9JKUVldn6RObvZkVN7EYutNl8GxRvYuRme3Motopm3DA.tOEGbuK	f	2026-01-22 20:41:53.741578+00	2026-01-11 16:25:35.373207+00	2026-01-22 20:41:53.741728+00
\.


--
-- Data for Name: votes; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.votes (id, poll_id, user_id, proposal_id, ticket_type, amount, created_at) FROM stdin;
8eec8435-7846-4e5f-9261-42ea706cfa82	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	daily_vote	1	2026-01-21 20:56:16.483273+00
e3de0376-c2f2-4ab2-88f4-df1c45aff7ca	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-21 21:29:24.288+00
b8be3e4d-26b3-492b-9e76-e0a05499a0af	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-21 21:29:24.973375+00
863179f1-948a-4362-99a0-52d935a8d0dd	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-21 21:29:25.149597+00
9e6da1a4-ce8c-4044-8630-7b7a5cd62b4d	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	daily_vote	1	2026-01-22 20:37:03.668682+00
41811167-e837-46f9-95e5-eab6904932a5	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-22 20:37:08.471216+00
b2085ac0-be13-452a-836b-ec17e2e30b6f	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-22 20:37:08.685437+00
5e30b6e9-b296-4317-ab71-472c1ae98f79	3868120d-127e-40f4-8f33-aa0140f8f80e	30fa06a3-de30-45ee-a148-f222c1da6e20	fd3b9c09-7f81-41c4-a459-3a2db58420f4	translation_ticket	1	2026-01-22 20:37:09.171516+00
\.


--
-- Data for Name: voting_polls; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.voting_polls (id, status, starts_at, ends_at, created_at) FROM stdin;
3868120d-127e-40f4-8f33-aa0140f8f80e	active	2026-01-21 20:56:16.475457+00	2026-01-22 20:56:16.475457+00	2026-01-21 20:56:16.475457+00
\.


--
-- Data for Name: weekly_ticket_grants; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.weekly_ticket_grants (id, grant_date, users_processed, novel_requests_granted, translation_tickets_granted, started_at, completed_at, status, error_message) FROM stdin;
34208f3d-ef30-4a14-9528-207b9e4d31b7	2026-01-21	2	0	0	2026-01-21 21:30:23.382556+00	2026-01-21 21:30:23.499978+00	completed	\N
\.


--
-- Data for Name: xp_events; Type: TABLE DATA; Schema: public; Owner: novels
--

COPY public.xp_events (id, user_id, type, delta, ref_type, ref_id, created_at) FROM stdin;
03f84d17-d3ba-489b-bbc1-14eee2ae4103	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	7c9c3a75-91be-4648-b180-5b2be0fa400e	2026-01-21 22:00:13.431879+00
914b5b59-e8c3-457a-8962-9e5d6b9d510a	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	222dc36d-e04c-4eb6-ba64-04fca8216eb5	2026-01-21 22:00:22.027751+00
4fb42dc2-a4ac-42da-80fa-42926ab48b61	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	aa6288ff-edae-45cd-8f6d-2e2a7577c06b	2026-01-21 22:02:58.888802+00
ca9b7aa2-484c-4cc1-b0b5-7abf50e59c3e	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	3b68fa60-e47e-416b-a308-d504a7b5bec3	2026-01-21 22:03:09.392282+00
c33431cf-6683-4336-8853-1e55e054fbc1	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	1b7ca659-177c-4a3d-95ef-e89df8c2ab0d	2026-01-21 22:11:33.969606+00
1e5efaea-5b45-415c-bf8b-b0f9d2bee996	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	59af45e7-1a76-4c0f-bca8-8ca91030a349	2026-01-21 22:11:38.566772+00
b592461d-9ea4-43a8-8840-3d945e2c4ebc	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	d8c53acf-7200-413f-a9e3-478fcd590b50	2026-01-21 22:11:58.493638+00
cbf240fe-dad6-4495-b6e9-d0e17e6a37c2	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	c8355372-d521-43b0-963e-ad3ea04d5800	2026-01-21 22:12:04.149481+00
4a33276c-23ed-4144-8a4a-9e7e79a0713a	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	8b389d1f-d0ac-4a3a-8883-e8dc372a7038	2026-01-22 20:20:39.036344+00
fea986f8-95b8-4879-b181-7f6e4d3c82fe	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	c6d79606-ea22-483b-ab72-2c6740adc300	2026-01-22 20:20:43.261798+00
a55c26ba-df7a-4f70-a023-c5e0094e5065	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	5db7d6cf-ed3a-40e2-94ce-525270df8682	2026-01-22 20:22:05.134973+00
87c43753-ca5d-49bd-ba76-4664d1a35d29	ce76495c-3f80-4c7a-be56-3d5994b49b03	comment	5	comment	25d1d476-e49a-4412-af42-3c030651693e	2026-01-22 20:35:00.836787+00
\.


--
-- Name: achievements achievements_code_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.achievements
    ADD CONSTRAINT achievements_code_key UNIQUE (code);


--
-- Name: achievements achievements_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.achievements
    ADD CONSTRAINT achievements_pkey PRIMARY KEY (id);


--
-- Name: admin_audit_log admin_audit_log_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.admin_audit_log
    ADD CONSTRAINT admin_audit_log_pkey PRIMARY KEY (id);


--
-- Name: app_settings app_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.app_settings
    ADD CONSTRAINT app_settings_pkey PRIMARY KEY (key);


--
-- Name: author_localizations author_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.author_localizations
    ADD CONSTRAINT author_localizations_pkey PRIMARY KEY (author_id, lang);


--
-- Name: authors authors_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_pkey PRIMARY KEY (id);


--
-- Name: authors authors_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.authors
    ADD CONSTRAINT authors_slug_key UNIQUE (slug);


--
-- Name: bookmark_lists bookmark_lists_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmark_lists
    ADD CONSTRAINT bookmark_lists_pkey PRIMARY KEY (id);


--
-- Name: bookmark_lists bookmark_lists_user_id_code_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmark_lists
    ADD CONSTRAINT bookmark_lists_user_id_code_key UNIQUE (user_id, code);


--
-- Name: bookmarks bookmarks_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_pkey PRIMARY KEY (user_id, novel_id);


--
-- Name: chapter_contents chapter_contents_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.chapter_contents
    ADD CONSTRAINT chapter_contents_pkey PRIMARY KEY (chapter_id, lang);


--
-- Name: chapters chapters_novel_id_number_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_novel_id_number_key UNIQUE (novel_id, number);


--
-- Name: chapters chapters_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_pkey PRIMARY KEY (id);


--
-- Name: collection_items collection_items_collection_id_novel_id_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT collection_items_collection_id_novel_id_key UNIQUE (collection_id, novel_id);


--
-- Name: collection_items collection_items_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT collection_items_pkey PRIMARY KEY (id);


--
-- Name: collection_votes collection_votes_collection_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_votes
    ADD CONSTRAINT collection_votes_collection_id_user_id_key UNIQUE (collection_id, user_id);


--
-- Name: collection_votes collection_votes_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_votes
    ADD CONSTRAINT collection_votes_pkey PRIMARY KEY (id);


--
-- Name: collections collections_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_pkey PRIMARY KEY (id);


--
-- Name: collections collections_user_id_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_user_id_slug_key UNIQUE (user_id, slug);


--
-- Name: comment_reports comment_reports_comment_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_comment_id_user_id_key UNIQUE (comment_id, user_id);


--
-- Name: comment_reports comment_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_pkey PRIMARY KEY (id);


--
-- Name: comment_votes comment_votes_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_votes
    ADD CONSTRAINT comment_votes_pkey PRIMARY KEY (comment_id, user_id);


--
-- Name: comments comments_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_pkey PRIMARY KEY (id);


--
-- Name: daily_vote_grants daily_vote_grants_grant_date_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.daily_vote_grants
    ADD CONSTRAINT daily_vote_grants_grant_date_key UNIQUE (grant_date);


--
-- Name: daily_vote_grants daily_vote_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.daily_vote_grants
    ADD CONSTRAINT daily_vote_grants_pkey PRIMARY KEY (id);


--
-- Name: genre_localizations genre_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.genre_localizations
    ADD CONSTRAINT genre_localizations_pkey PRIMARY KEY (genre_id, lang);


--
-- Name: genres genres_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_pkey PRIMARY KEY (id);


--
-- Name: genres genres_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_slug_key UNIQUE (slug);


--
-- Name: leaderboard_cache leaderboard_cache_period_user_id_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.leaderboard_cache
    ADD CONSTRAINT leaderboard_cache_period_user_id_key UNIQUE (period, user_id);


--
-- Name: leaderboard_cache leaderboard_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.leaderboard_cache
    ADD CONSTRAINT leaderboard_cache_pkey PRIMARY KEY (id);


--
-- Name: news_localizations news_localizations_news_id_lang_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_localizations
    ADD CONSTRAINT news_localizations_news_id_lang_key UNIQUE (news_id, lang);


--
-- Name: news_localizations news_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_localizations
    ADD CONSTRAINT news_localizations_pkey PRIMARY KEY (id);


--
-- Name: news_posts news_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_posts
    ADD CONSTRAINT news_posts_pkey PRIMARY KEY (id);


--
-- Name: news_posts news_posts_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_posts
    ADD CONSTRAINT news_posts_slug_key UNIQUE (slug);


--
-- Name: novel_authors novel_authors_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_authors
    ADD CONSTRAINT novel_authors_pkey PRIMARY KEY (novel_id, author_id);


--
-- Name: novel_edit_history novel_edit_history_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_history
    ADD CONSTRAINT novel_edit_history_pkey PRIMARY KEY (id);


--
-- Name: novel_edit_request_changes novel_edit_request_changes_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_request_changes
    ADD CONSTRAINT novel_edit_request_changes_pkey PRIMARY KEY (id);


--
-- Name: novel_edit_requests novel_edit_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_requests
    ADD CONSTRAINT novel_edit_requests_pkey PRIMARY KEY (id);


--
-- Name: novel_genres novel_genres_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_genres
    ADD CONSTRAINT novel_genres_pkey PRIMARY KEY (novel_id, genre_id);


--
-- Name: novel_localizations novel_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_localizations
    ADD CONSTRAINT novel_localizations_pkey PRIMARY KEY (novel_id, lang);


--
-- Name: novel_proposals novel_proposals_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_proposals
    ADD CONSTRAINT novel_proposals_pkey PRIMARY KEY (id);


--
-- Name: novel_ratings novel_ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_ratings
    ADD CONSTRAINT novel_ratings_pkey PRIMARY KEY (novel_id, user_id);


--
-- Name: novel_tags novel_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_tags
    ADD CONSTRAINT novel_tags_pkey PRIMARY KEY (novel_id, tag_id);


--
-- Name: novel_views_daily novel_views_daily_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_views_daily
    ADD CONSTRAINT novel_views_daily_pkey PRIMARY KEY (novel_id, date);


--
-- Name: novels novels_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novels
    ADD CONSTRAINT novels_pkey PRIMARY KEY (id);


--
-- Name: novels novels_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novels
    ADD CONSTRAINT novels_slug_key UNIQUE (slug);


--
-- Name: platform_stats platform_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.platform_stats
    ADD CONSTRAINT platform_stats_pkey PRIMARY KEY (id);


--
-- Name: reading_progress reading_progress_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.reading_progress
    ADD CONSTRAINT reading_progress_pkey PRIMARY KEY (user_id, novel_id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: subscription_grants subscription_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_grants
    ADD CONSTRAINT subscription_grants_pkey PRIMARY KEY (id);


--
-- Name: subscription_grants subscription_grants_subscription_id_type_for_month_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_grants
    ADD CONSTRAINT subscription_grants_subscription_id_type_for_month_key UNIQUE (subscription_id, type, for_month);


--
-- Name: subscription_plans subscription_plans_code_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_plans
    ADD CONSTRAINT subscription_plans_code_key UNIQUE (code);


--
-- Name: subscription_plans subscription_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_plans
    ADD CONSTRAINT subscription_plans_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: tag_localizations tag_localizations_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.tag_localizations
    ADD CONSTRAINT tag_localizations_pkey PRIMARY KEY (tag_id, lang);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: tags tags_slug_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_slug_key UNIQUE (slug);


--
-- Name: ticket_balances ticket_balances_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.ticket_balances
    ADD CONSTRAINT ticket_balances_pkey PRIMARY KEY (user_id, type);


--
-- Name: ticket_transactions ticket_transactions_idempotency_key_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.ticket_transactions
    ADD CONSTRAINT ticket_transactions_idempotency_key_key UNIQUE (idempotency_key);


--
-- Name: ticket_transactions ticket_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.ticket_transactions
    ADD CONSTRAINT ticket_transactions_pkey PRIMARY KEY (id);


--
-- Name: user_achievements user_achievements_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_achievements
    ADD CONSTRAINT user_achievements_pkey PRIMARY KEY (user_id, achievement_id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (user_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role);


--
-- Name: user_xp user_xp_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_xp
    ADD CONSTRAINT user_xp_pkey PRIMARY KEY (user_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: votes votes_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_pkey PRIMARY KEY (id);


--
-- Name: voting_polls voting_polls_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.voting_polls
    ADD CONSTRAINT voting_polls_pkey PRIMARY KEY (id);


--
-- Name: weekly_ticket_grants weekly_ticket_grants_grant_date_key; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.weekly_ticket_grants
    ADD CONSTRAINT weekly_ticket_grants_grant_date_key UNIQUE (grant_date);


--
-- Name: weekly_ticket_grants weekly_ticket_grants_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.weekly_ticket_grants
    ADD CONSTRAINT weekly_ticket_grants_pkey PRIMARY KEY (id);


--
-- Name: xp_events xp_events_pkey; Type: CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.xp_events
    ADD CONSTRAINT xp_events_pkey PRIMARY KEY (id);


--
-- Name: idx_admin_audit_log_action; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_admin_audit_log_action ON public.admin_audit_log USING btree (action);


--
-- Name: idx_admin_audit_log_actor; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_admin_audit_log_actor ON public.admin_audit_log USING btree (actor_user_id);


--
-- Name: idx_admin_audit_log_created_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_admin_audit_log_created_at ON public.admin_audit_log USING btree (created_at DESC);


--
-- Name: idx_admin_audit_log_entity; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_admin_audit_log_entity ON public.admin_audit_log USING btree (entity_type, entity_id);


--
-- Name: idx_author_localizations_name; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_author_localizations_name ON public.author_localizations USING btree (name);


--
-- Name: idx_authors_slug; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_authors_slug ON public.authors USING btree (slug);


--
-- Name: idx_bookmark_lists_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmark_lists_user ON public.bookmark_lists USING btree (user_id);


--
-- Name: idx_bookmarks_list; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_list ON public.bookmarks USING btree (list_id);


--
-- Name: idx_bookmarks_list_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_list_id ON public.bookmarks USING btree (list_id);


--
-- Name: idx_bookmarks_novel; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_novel ON public.bookmarks USING btree (novel_id);


--
-- Name: idx_bookmarks_updated_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_updated_at ON public.bookmarks USING btree (updated_at DESC);


--
-- Name: idx_bookmarks_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_user ON public.bookmarks USING btree (user_id);


--
-- Name: idx_bookmarks_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_bookmarks_user_id ON public.bookmarks USING btree (user_id);


--
-- Name: idx_chapters_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_chapters_novel_id ON public.chapters USING btree (novel_id);


--
-- Name: idx_chapters_number; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_chapters_number ON public.chapters USING btree (novel_id, number);


--
-- Name: idx_chapters_published_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_chapters_published_at ON public.chapters USING btree (published_at DESC);


--
-- Name: idx_collection_items_collection_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collection_items_collection_id ON public.collection_items USING btree (collection_id);


--
-- Name: idx_collection_items_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collection_items_novel_id ON public.collection_items USING btree (novel_id);


--
-- Name: idx_collection_items_position; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collection_items_position ON public.collection_items USING btree (collection_id, "position");


--
-- Name: idx_collection_votes_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collection_votes_user_id ON public.collection_votes USING btree (user_id);


--
-- Name: idx_collections_created_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collections_created_at ON public.collections USING btree (created_at DESC);


--
-- Name: idx_collections_is_featured; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collections_is_featured ON public.collections USING btree (is_featured) WHERE (is_featured = true);


--
-- Name: idx_collections_is_public; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collections_is_public ON public.collections USING btree (is_public) WHERE (is_public = true);


--
-- Name: idx_collections_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collections_user_id ON public.collections USING btree (user_id);


--
-- Name: idx_collections_votes_count; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_collections_votes_count ON public.collections USING btree (votes_count DESC);


--
-- Name: idx_comment_reports_comment; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comment_reports_comment ON public.comment_reports USING btree (comment_id);


--
-- Name: idx_comment_reports_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comment_reports_status ON public.comment_reports USING btree (status);


--
-- Name: idx_comment_votes_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comment_votes_user ON public.comment_votes USING btree (user_id);


--
-- Name: idx_comments_chapter_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_chapter_id ON public.comments USING btree (chapter_id);


--
-- Name: idx_comments_created_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_created_at ON public.comments USING btree (created_at DESC);


--
-- Name: idx_comments_depth; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_depth ON public.comments USING btree (depth);


--
-- Name: idx_comments_is_deleted; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_is_deleted ON public.comments USING btree (is_deleted) WHERE (is_deleted = true);


--
-- Name: idx_comments_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_novel_id ON public.comments USING btree (novel_id);


--
-- Name: idx_comments_parent_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_parent_id ON public.comments USING btree (parent_id);


--
-- Name: idx_comments_root_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_root_id ON public.comments USING btree (root_id);


--
-- Name: idx_comments_target; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_target ON public.comments USING btree (target_type, target_id);


--
-- Name: idx_comments_target_anchor; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_target_anchor ON public.comments USING btree (target_type, target_id, anchor);


--
-- Name: idx_comments_target_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_target_id ON public.comments USING btree (target_id);


--
-- Name: idx_comments_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_comments_user_id ON public.comments USING btree (user_id);


--
-- Name: idx_daily_vote_grants_date; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_daily_vote_grants_date ON public.daily_vote_grants USING btree (grant_date);


--
-- Name: idx_edit_history_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_history_novel_id ON public.novel_edit_history USING btree (novel_id);


--
-- Name: idx_edit_history_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_history_user_id ON public.novel_edit_history USING btree (user_id);


--
-- Name: idx_edit_requests_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_requests_novel_id ON public.novel_edit_requests USING btree (novel_id);


--
-- Name: idx_edit_requests_pending; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_requests_pending ON public.novel_edit_requests USING btree (status, created_at) WHERE (status = 'pending'::public.edit_request_status);


--
-- Name: idx_edit_requests_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_requests_status ON public.novel_edit_requests USING btree (status);


--
-- Name: idx_edit_requests_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_edit_requests_user_id ON public.novel_edit_requests USING btree (user_id);


--
-- Name: idx_leaderboard_cache_period; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_leaderboard_cache_period ON public.leaderboard_cache USING btree (period, rank);


--
-- Name: idx_leaderboard_cache_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_leaderboard_cache_user ON public.leaderboard_cache USING btree (user_id);


--
-- Name: idx_news_localizations_news_lang; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_localizations_news_lang ON public.news_localizations USING btree (news_id, lang);


--
-- Name: idx_news_posts_author; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_author ON public.news_posts USING btree (author_id);


--
-- Name: idx_news_posts_category; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_category ON public.news_posts USING btree (category);


--
-- Name: idx_news_posts_is_pinned; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_is_pinned ON public.news_posts USING btree (is_pinned) WHERE (is_pinned = true);


--
-- Name: idx_news_posts_is_published; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_is_published ON public.news_posts USING btree (is_published) WHERE (is_published = true);


--
-- Name: idx_news_posts_published_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_published_at ON public.news_posts USING btree (published_at DESC);


--
-- Name: idx_news_posts_slug; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_news_posts_slug ON public.news_posts USING btree (slug);


--
-- Name: idx_novel_authors_author_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_authors_author_id ON public.novel_authors USING btree (author_id);


--
-- Name: idx_novel_authors_is_primary; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_authors_is_primary ON public.novel_authors USING btree (is_primary);


--
-- Name: idx_novel_authors_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_authors_novel_id ON public.novel_authors USING btree (novel_id);


--
-- Name: idx_novel_localizations_search; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_localizations_search ON public.novel_localizations USING gin (search_vector);


--
-- Name: idx_novel_localizations_title; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_localizations_title ON public.novel_localizations USING gin (title public.gin_trgm_ops);


--
-- Name: idx_novel_proposals_created; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_proposals_created ON public.novel_proposals USING btree (created_at);


--
-- Name: idx_novel_proposals_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_proposals_status ON public.novel_proposals USING btree (status);


--
-- Name: idx_novel_proposals_translation_invested; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_proposals_translation_invested ON public.novel_proposals USING btree (translation_tickets_invested DESC);


--
-- Name: idx_novel_proposals_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_proposals_user ON public.novel_proposals USING btree (user_id);


--
-- Name: idx_novel_proposals_voting; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_proposals_voting ON public.novel_proposals USING btree (status, vote_score DESC) WHERE (status = 'voting'::public.proposal_status);


--
-- Name: idx_novel_views_daily_date; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novel_views_daily_date ON public.novel_views_daily USING btree (date DESC);


--
-- Name: idx_novels_created_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_created_at ON public.novels USING btree (created_at DESC);


--
-- Name: idx_novels_rating; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_rating ON public.novels USING btree ((((rating_sum)::double precision / (NULLIF(rating_count, 0))::double precision)) DESC NULLS LAST);


--
-- Name: idx_novels_slug; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_slug ON public.novels USING btree (slug);


--
-- Name: idx_novels_translation_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_translation_status ON public.novels USING btree (translation_status);


--
-- Name: idx_novels_updated_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_updated_at ON public.novels USING btree (updated_at DESC);


--
-- Name: idx_novels_views_daily; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_views_daily ON public.novels USING btree (views_daily DESC);


--
-- Name: idx_novels_views_total; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_novels_views_total ON public.novels USING btree (views_total DESC);


--
-- Name: idx_reading_progress_novel_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_reading_progress_novel_id ON public.reading_progress USING btree (novel_id);


--
-- Name: idx_reading_progress_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_reading_progress_user_id ON public.reading_progress USING btree (user_id);


--
-- Name: idx_refresh_tokens_expires_at; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_refresh_tokens_expires_at ON public.refresh_tokens USING btree (expires_at);


--
-- Name: idx_refresh_tokens_user_id; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_subscription_grants_month; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscription_grants_month ON public.subscription_grants USING btree (for_month);


--
-- Name: idx_subscription_grants_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscription_grants_user ON public.subscription_grants USING btree (user_id);


--
-- Name: idx_subscriptions_active; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscriptions_active ON public.subscriptions USING btree (user_id, ends_at) WHERE (status = 'active'::public.subscription_status);


--
-- Name: idx_subscriptions_expires; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscriptions_expires ON public.subscriptions USING btree (ends_at) WHERE (status = 'active'::public.subscription_status);


--
-- Name: idx_subscriptions_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscriptions_status ON public.subscriptions USING btree (status);


--
-- Name: idx_subscriptions_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_subscriptions_user ON public.subscriptions USING btree (user_id);


--
-- Name: idx_ticket_balances_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_ticket_balances_user ON public.ticket_balances USING btree (user_id);


--
-- Name: idx_ticket_transactions_created; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_ticket_transactions_created ON public.ticket_transactions USING btree (created_at);


--
-- Name: idx_ticket_transactions_idempotency; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_ticket_transactions_idempotency ON public.ticket_transactions USING btree (idempotency_key) WHERE (idempotency_key IS NOT NULL);


--
-- Name: idx_ticket_transactions_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_ticket_transactions_user ON public.ticket_transactions USING btree (user_id);


--
-- Name: idx_ticket_transactions_user_type; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_ticket_transactions_user_type ON public.ticket_transactions USING btree (user_id, type);


--
-- Name: idx_user_achievements_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_user_achievements_user ON public.user_achievements USING btree (user_id);


--
-- Name: idx_user_xp_level; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_user_xp_level ON public.user_xp USING btree (level);


--
-- Name: idx_user_xp_total; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_user_xp_total ON public.user_xp USING btree (xp_total DESC);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_votes_created; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_votes_created ON public.votes USING btree (created_at);


--
-- Name: idx_votes_poll; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_votes_poll ON public.votes USING btree (poll_id);


--
-- Name: idx_votes_proposal; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_votes_proposal ON public.votes USING btree (proposal_id);


--
-- Name: idx_votes_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_votes_user ON public.votes USING btree (user_id);


--
-- Name: idx_voting_polls_active; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_voting_polls_active ON public.voting_polls USING btree (ends_at) WHERE ((status)::text = 'active'::text);


--
-- Name: idx_voting_polls_status; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_voting_polls_status ON public.voting_polls USING btree (status);


--
-- Name: idx_weekly_ticket_grants_date; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_weekly_ticket_grants_date ON public.weekly_ticket_grants USING btree (grant_date);


--
-- Name: idx_xp_events_created; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_xp_events_created ON public.xp_events USING btree (created_at);


--
-- Name: idx_xp_events_type; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_xp_events_type ON public.xp_events USING btree (type);


--
-- Name: idx_xp_events_user; Type: INDEX; Schema: public; Owner: novels
--

CREATE INDEX idx_xp_events_user ON public.xp_events USING btree (user_id);


--
-- Name: votes trg_update_proposal_stats; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trg_update_proposal_stats AFTER INSERT OR DELETE ON public.votes FOR EACH ROW EXECUTE FUNCTION public.update_proposal_stats();


--
-- Name: novel_proposals trg_update_proposal_timestamp; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trg_update_proposal_timestamp BEFORE UPDATE ON public.novel_proposals FOR EACH ROW EXECUTE FUNCTION public.update_subscription_timestamp();


--
-- Name: subscriptions trg_update_subscription_timestamp; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trg_update_subscription_timestamp BEFORE UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.update_subscription_timestamp();


--
-- Name: ticket_transactions trg_update_ticket_balance; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trg_update_ticket_balance AFTER INSERT ON public.ticket_transactions FOR EACH ROW EXECUTE FUNCTION public.update_ticket_balance();


--
-- Name: collection_items trigger_collection_items_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_collection_items_count AFTER INSERT OR DELETE ON public.collection_items FOR EACH ROW EXECUTE FUNCTION public.update_collection_items_count();


--
-- Name: collection_votes trigger_collection_votes_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_collection_votes_count AFTER INSERT OR DELETE ON public.collection_votes FOR EACH ROW EXECUTE FUNCTION public.update_collection_votes_count();


--
-- Name: collections trigger_collections_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_collections_updated_at BEFORE UPDATE ON public.collections FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: comments trigger_comment_replies_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_comment_replies_count AFTER INSERT OR DELETE OR UPDATE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_comment_replies_count();


--
-- Name: comment_votes trigger_comment_vote_counts; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_comment_vote_counts AFTER INSERT OR DELETE OR UPDATE ON public.comment_votes FOR EACH ROW EXECUTE FUNCTION public.update_comment_vote_counts();


--
-- Name: novel_edit_requests trigger_edit_requests_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_edit_requests_updated_at BEFORE UPDATE ON public.novel_edit_requests FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: comments trigger_news_comments_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_news_comments_count AFTER INSERT OR DELETE ON public.comments FOR EACH ROW EXECUTE FUNCTION public.update_news_comments_count();


--
-- Name: news_posts trigger_news_posts_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_news_posts_updated_at BEFORE UPDATE ON public.news_posts FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: comments trigger_set_comment_hierarchy; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER trigger_set_comment_hierarchy BEFORE INSERT ON public.comments FOR EACH ROW EXECUTE FUNCTION public.set_comment_hierarchy();


--
-- Name: app_settings update_app_settings_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_app_settings_updated_at BEFORE UPDATE ON public.app_settings FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: author_localizations update_author_localizations_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_author_localizations_updated_at BEFORE UPDATE ON public.author_localizations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: authors update_authors_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_authors_updated_at BEFORE UPDATE ON public.authors FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: bookmarks update_bookmarks_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_bookmarks_updated_at BEFORE UPDATE ON public.bookmarks FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: chapter_contents update_chapter_contents_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_chapter_contents_updated_at BEFORE UPDATE ON public.chapter_contents FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: chapter_contents update_chapter_contents_word_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_chapter_contents_word_count BEFORE INSERT OR UPDATE ON public.chapter_contents FOR EACH ROW EXECUTE FUNCTION public.update_chapter_word_count();


--
-- Name: chapters update_chapters_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_chapters_updated_at BEFORE UPDATE ON public.chapters FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: novel_localizations update_novel_localizations_search_vector; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_novel_localizations_search_vector BEFORE INSERT OR UPDATE ON public.novel_localizations FOR EACH ROW EXECUTE FUNCTION public.update_novel_search_vector();


--
-- Name: novel_localizations update_novel_localizations_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_novel_localizations_updated_at BEFORE UPDATE ON public.novel_localizations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: bookmarks update_novels_bookmarks_count; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_novels_bookmarks_count AFTER INSERT OR DELETE ON public.bookmarks FOR EACH ROW EXECUTE FUNCTION public.update_novel_bookmarks_count();


--
-- Name: novel_ratings update_novels_rating; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_novels_rating AFTER INSERT OR DELETE OR UPDATE ON public.novel_ratings FOR EACH ROW EXECUTE FUNCTION public.update_novel_rating();


--
-- Name: novels update_novels_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_novels_updated_at BEFORE UPDATE ON public.novels FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: user_profiles update_user_profiles_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON public.user_profiles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: novels
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: admin_audit_log admin_audit_log_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.admin_audit_log
    ADD CONSTRAINT admin_audit_log_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: app_settings app_settings_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.app_settings
    ADD CONSTRAINT app_settings_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: author_localizations author_localizations_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.author_localizations
    ADD CONSTRAINT author_localizations_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.authors(id) ON DELETE CASCADE;


--
-- Name: bookmark_lists bookmark_lists_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmark_lists
    ADD CONSTRAINT bookmark_lists_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: bookmarks bookmarks_list_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_list_id_fkey FOREIGN KEY (list_id) REFERENCES public.bookmark_lists(id) ON DELETE CASCADE;


--
-- Name: bookmarks bookmarks_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: bookmarks bookmarks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: chapter_contents chapter_contents_chapter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.chapter_contents
    ADD CONSTRAINT chapter_contents_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES public.chapters(id) ON DELETE CASCADE;


--
-- Name: chapters chapters_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.chapters
    ADD CONSTRAINT chapters_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: collection_items collection_items_collection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT collection_items_collection_id_fkey FOREIGN KEY (collection_id) REFERENCES public.collections(id) ON DELETE CASCADE;


--
-- Name: collection_items collection_items_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_items
    ADD CONSTRAINT collection_items_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: collection_votes collection_votes_collection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_votes
    ADD CONSTRAINT collection_votes_collection_id_fkey FOREIGN KEY (collection_id) REFERENCES public.collections(id) ON DELETE CASCADE;


--
-- Name: collection_votes collection_votes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collection_votes
    ADD CONSTRAINT collection_votes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: collections collections_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comment_reports comment_reports_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comment_reports comment_reports_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_reports
    ADD CONSTRAINT comment_reports_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comment_votes comment_votes_comment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_votes
    ADD CONSTRAINT comment_votes_comment_id_fkey FOREIGN KEY (comment_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comment_votes comment_votes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comment_votes
    ADD CONSTRAINT comment_votes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: comments comments_chapter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES public.chapters(id) ON DELETE CASCADE;


--
-- Name: comments comments_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: comments comments_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comments comments_root_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_root_id_fkey FOREIGN KEY (root_id) REFERENCES public.comments(id) ON DELETE CASCADE;


--
-- Name: comments comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.comments
    ADD CONSTRAINT comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: genre_localizations genre_localizations_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.genre_localizations
    ADD CONSTRAINT genre_localizations_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;


--
-- Name: leaderboard_cache leaderboard_cache_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.leaderboard_cache
    ADD CONSTRAINT leaderboard_cache_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: news_localizations news_localizations_news_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_localizations
    ADD CONSTRAINT news_localizations_news_id_fkey FOREIGN KEY (news_id) REFERENCES public.news_posts(id) ON DELETE CASCADE;


--
-- Name: news_posts news_posts_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.news_posts
    ADD CONSTRAINT news_posts_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.users(id);


--
-- Name: novel_authors novel_authors_author_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_authors
    ADD CONSTRAINT novel_authors_author_id_fkey FOREIGN KEY (author_id) REFERENCES public.authors(id) ON DELETE CASCADE;


--
-- Name: novel_authors novel_authors_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_authors
    ADD CONSTRAINT novel_authors_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_edit_history novel_edit_history_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_history
    ADD CONSTRAINT novel_edit_history_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_edit_history novel_edit_history_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_history
    ADD CONSTRAINT novel_edit_history_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.novel_edit_requests(id);


--
-- Name: novel_edit_history novel_edit_history_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_history
    ADD CONSTRAINT novel_edit_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: novel_edit_request_changes novel_edit_request_changes_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_request_changes
    ADD CONSTRAINT novel_edit_request_changes_request_id_fkey FOREIGN KEY (request_id) REFERENCES public.novel_edit_requests(id) ON DELETE CASCADE;


--
-- Name: novel_edit_requests novel_edit_requests_moderator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_requests
    ADD CONSTRAINT novel_edit_requests_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id);


--
-- Name: novel_edit_requests novel_edit_requests_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_requests
    ADD CONSTRAINT novel_edit_requests_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_edit_requests novel_edit_requests_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_edit_requests
    ADD CONSTRAINT novel_edit_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: novel_genres novel_genres_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_genres
    ADD CONSTRAINT novel_genres_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;


--
-- Name: novel_genres novel_genres_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_genres
    ADD CONSTRAINT novel_genres_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_localizations novel_localizations_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_localizations
    ADD CONSTRAINT novel_localizations_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_proposals novel_proposals_moderator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_proposals
    ADD CONSTRAINT novel_proposals_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id);


--
-- Name: novel_proposals novel_proposals_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_proposals
    ADD CONSTRAINT novel_proposals_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: novel_ratings novel_ratings_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_ratings
    ADD CONSTRAINT novel_ratings_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_ratings novel_ratings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_ratings
    ADD CONSTRAINT novel_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: novel_tags novel_tags_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_tags
    ADD CONSTRAINT novel_tags_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: novel_tags novel_tags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_tags
    ADD CONSTRAINT novel_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: novel_views_daily novel_views_daily_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.novel_views_daily
    ADD CONSTRAINT novel_views_daily_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: reading_progress reading_progress_chapter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.reading_progress
    ADD CONSTRAINT reading_progress_chapter_id_fkey FOREIGN KEY (chapter_id) REFERENCES public.chapters(id) ON DELETE CASCADE;


--
-- Name: reading_progress reading_progress_novel_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.reading_progress
    ADD CONSTRAINT reading_progress_novel_id_fkey FOREIGN KEY (novel_id) REFERENCES public.novels(id) ON DELETE CASCADE;


--
-- Name: reading_progress reading_progress_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.reading_progress
    ADD CONSTRAINT reading_progress_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: subscription_grants subscription_grants_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_grants
    ADD CONSTRAINT subscription_grants_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE CASCADE;


--
-- Name: subscription_grants subscription_grants_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscription_grants
    ADD CONSTRAINT subscription_grants_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.subscription_plans(id);


--
-- Name: subscriptions subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: tag_localizations tag_localizations_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.tag_localizations
    ADD CONSTRAINT tag_localizations_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id) ON DELETE CASCADE;


--
-- Name: ticket_balances ticket_balances_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.ticket_balances
    ADD CONSTRAINT ticket_balances_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: ticket_transactions ticket_transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.ticket_transactions
    ADD CONSTRAINT ticket_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_achievements user_achievements_achievement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_achievements
    ADD CONSTRAINT user_achievements_achievement_id_fkey FOREIGN KEY (achievement_id) REFERENCES public.achievements(id) ON DELETE CASCADE;


--
-- Name: user_achievements user_achievements_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_achievements
    ADD CONSTRAINT user_achievements_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_profiles user_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_xp user_xp_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.user_xp
    ADD CONSTRAINT user_xp_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: votes votes_poll_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_poll_id_fkey FOREIGN KEY (poll_id) REFERENCES public.voting_polls(id) ON DELETE SET NULL;


--
-- Name: votes votes_proposal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_proposal_id_fkey FOREIGN KEY (proposal_id) REFERENCES public.novel_proposals(id) ON DELETE CASCADE;


--
-- Name: votes votes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.votes
    ADD CONSTRAINT votes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: xp_events xp_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: novels
--

ALTER TABLE ONLY public.xp_events
    ADD CONSTRAINT xp_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

