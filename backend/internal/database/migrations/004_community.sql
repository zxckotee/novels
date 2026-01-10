-- Migration 004: Community features (collections, news, wiki edits)
-- Пользовательские коллекции, новости платформы, система редактирования

-- ================================================
-- COLLECTIONS (Коллекции пользователей)
-- ================================================

-- Коллекции (подборки книг)
CREATE TABLE IF NOT EXISTS collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_url TEXT,
    is_public BOOLEAN DEFAULT true,
    is_featured BOOLEAN DEFAULT false,
    views_count INTEGER DEFAULT 0,
    votes_count INTEGER DEFAULT 0,
    items_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, slug)
);

-- Элементы коллекции
CREATE TABLE IF NOT EXISTS collection_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0,
    note TEXT, -- Заметка пользователя о книге в коллекции
    added_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(collection_id, novel_id)
);

-- Голоса за коллекции
CREATE TABLE IF NOT EXISTS collection_votes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL DEFAULT 1, -- +1
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(collection_id, user_id)
);

-- ================================================
-- NEWS (Новости платформы)
-- ================================================

-- Категории новостей
CREATE TYPE news_category AS ENUM (
    'announcement',    -- Объявления
    'update',         -- Обновления платформы
    'event',          -- События
    'community',      -- Новости сообщества
    'translation'     -- Новости переводов
);

-- Новости
CREATE TABLE IF NOT EXISTS news_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    summary TEXT, -- Краткое описание для превью
    content TEXT NOT NULL, -- Полный текст (HTML/Markdown)
    cover_url TEXT,
    category news_category DEFAULT 'announcement',
    author_id UUID NOT NULL REFERENCES users(id),
    is_published BOOLEAN DEFAULT false,
    is_pinned BOOLEAN DEFAULT false,
    views_count INTEGER DEFAULT 0,
    comments_count INTEGER DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Локализации новостей
CREATE TABLE IF NOT EXISTS news_localizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID NOT NULL REFERENCES news_posts(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(news_id, lang)
);

-- ================================================
-- WIKI EDITS (Редактирование описаний через модерацию)
-- ================================================

-- Статусы запросов на редактирование
CREATE TYPE edit_request_status AS ENUM (
    'pending',        -- Ожидает модерации
    'approved',       -- Одобрено и применено
    'rejected',       -- Отклонено
    'withdrawn'       -- Отозвано автором
);

-- Типы редактируемых полей
CREATE TYPE edit_field_type AS ENUM (
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

-- Запросы на редактирование новелл
CREATE TABLE IF NOT EXISTS novel_edit_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status edit_request_status DEFAULT 'pending',
    edit_reason TEXT, -- Почему пользователь хочет изменить
    moderator_id UUID REFERENCES users(id),
    moderator_comment TEXT, -- Комментарий модератора при отклонении
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Отдельные изменения в запросе (diff)
CREATE TABLE IF NOT EXISTS novel_edit_request_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES novel_edit_requests(id) ON DELETE CASCADE,
    field_type edit_field_type NOT NULL,
    lang VARCHAR(10), -- NULL для non-localized полей
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- История принятых изменений (аудит)
CREATE TABLE IF NOT EXISTS novel_edit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    request_id UUID REFERENCES novel_edit_requests(id), -- NULL если изменено админом напрямую
    user_id UUID NOT NULL REFERENCES users(id),
    field_type edit_field_type NOT NULL,
    lang VARCHAR(10),
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ================================================
-- GLOBAL STATISTICS (Глобальная статистика)
-- ================================================

-- Кеш глобальной статистики (обновляется периодически)
CREATE TABLE IF NOT EXISTS platform_stats (
    id INTEGER PRIMARY KEY DEFAULT 1,
    total_novels INTEGER DEFAULT 0,
    total_chapters INTEGER DEFAULT 0,
    total_users INTEGER DEFAULT 0,
    total_comments INTEGER DEFAULT 0,
    total_collections INTEGER DEFAULT 0,
    total_votes_cast INTEGER DEFAULT 0,
    total_tickets_spent BIGINT DEFAULT 0,
    proposals_translated INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT single_row CHECK (id = 1)
);

-- Инициализация статистики
INSERT INTO platform_stats (id) VALUES (1) ON CONFLICT DO NOTHING;

-- ================================================
-- INDEXES
-- ================================================

-- Collections indexes
CREATE INDEX idx_collections_user_id ON collections(user_id);
CREATE INDEX idx_collections_is_public ON collections(is_public) WHERE is_public = true;
CREATE INDEX idx_collections_is_featured ON collections(is_featured) WHERE is_featured = true;
CREATE INDEX idx_collections_votes_count ON collections(votes_count DESC);
CREATE INDEX idx_collections_created_at ON collections(created_at DESC);

CREATE INDEX idx_collection_items_collection_id ON collection_items(collection_id);
CREATE INDEX idx_collection_items_novel_id ON collection_items(novel_id);
CREATE INDEX idx_collection_items_position ON collection_items(collection_id, position);

CREATE INDEX idx_collection_votes_user_id ON collection_votes(user_id);

-- News indexes
CREATE INDEX idx_news_posts_slug ON news_posts(slug);
CREATE INDEX idx_news_posts_category ON news_posts(category);
CREATE INDEX idx_news_posts_is_published ON news_posts(is_published) WHERE is_published = true;
CREATE INDEX idx_news_posts_is_pinned ON news_posts(is_pinned) WHERE is_pinned = true;
CREATE INDEX idx_news_posts_published_at ON news_posts(published_at DESC);
CREATE INDEX idx_news_posts_author ON news_posts(author_id);

CREATE INDEX idx_news_localizations_news_lang ON news_localizations(news_id, lang);

-- Edit requests indexes
CREATE INDEX idx_edit_requests_novel_id ON novel_edit_requests(novel_id);
CREATE INDEX idx_edit_requests_user_id ON novel_edit_requests(user_id);
CREATE INDEX idx_edit_requests_status ON novel_edit_requests(status);
CREATE INDEX idx_edit_requests_pending ON novel_edit_requests(status, created_at) WHERE status = 'pending';

CREATE INDEX idx_edit_history_novel_id ON novel_edit_history(novel_id);
CREATE INDEX idx_edit_history_user_id ON novel_edit_history(user_id);

-- ================================================
-- TRIGGERS
-- ================================================

-- Trigger: автообновление items_count в коллекции
CREATE OR REPLACE FUNCTION update_collection_items_count()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_collection_items_count
AFTER INSERT OR DELETE ON collection_items
FOR EACH ROW EXECUTE FUNCTION update_collection_items_count();

-- Trigger: автообновление votes_count в коллекции
CREATE OR REPLACE FUNCTION update_collection_votes_count()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_collection_votes_count
AFTER INSERT OR DELETE ON collection_votes
FOR EACH ROW EXECUTE FUNCTION update_collection_votes_count();

-- Trigger: автообновление comments_count в новости
CREATE OR REPLACE FUNCTION update_news_comments_count()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

-- Сначала добавим новый тип в target_type если нужно
DO $$
BEGIN
    -- Проверяем существует ли значение 'news' в enum
    IF NOT EXISTS (
        SELECT 1 FROM pg_enum 
        WHERE enumlabel = 'news' 
        AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'comment_target_type')
    ) THEN
        ALTER TYPE comment_target_type ADD VALUE IF NOT EXISTS 'news';
    END IF;
EXCEPTION
    WHEN others THEN
        NULL;
END $$;

CREATE TRIGGER trigger_news_comments_count
AFTER INSERT OR DELETE ON comments
FOR EACH ROW EXECUTE FUNCTION update_news_comments_count();

-- Trigger: обновление updated_at
CREATE TRIGGER trigger_collections_updated_at
BEFORE UPDATE ON collections
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_news_posts_updated_at
BEFORE UPDATE ON news_posts
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_edit_requests_updated_at
BEFORE UPDATE ON novel_edit_requests
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ================================================
-- FUNCTIONS
-- ================================================

-- Функция: применить изменения из запроса на редактирование
CREATE OR REPLACE FUNCTION apply_edit_request(p_request_id UUID, p_moderator_id UUID)
RETURNS BOOLEAN AS $$
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
$$ LANGUAGE plpgsql;

-- Функция: обновление глобальной статистики
CREATE OR REPLACE FUNCTION refresh_platform_stats()
RETURNS VOID AS $$
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
$$ LANGUAGE plpgsql;
