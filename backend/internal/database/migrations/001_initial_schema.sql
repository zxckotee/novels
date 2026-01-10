-- Расширения PostgreSQL
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- Для полнотекстового поиска

-- Типы данных (Enums)
CREATE TYPE user_role AS ENUM ('guest', 'user', 'premium', 'moderator', 'admin');
CREATE TYPE translation_status AS ENUM ('ongoing', 'completed', 'paused', 'dropped');
CREATE TYPE bookmark_list_code AS ENUM ('reading', 'planned', 'dropped', 'completed', 'favorites');

-- ============================================
-- ПОЛЬЗОВАТЕЛИ
-- ============================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_banned BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100) NOT NULL,
    avatar_key VARCHAR(255),
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role)
);

-- Refresh токены для JWT
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- ============================================
-- ЖАНРЫ И ТЕГИ
-- ============================================

CREATE TABLE genres (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE genre_localizations (
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (genre_id, lang)
);

CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE tag_localizations (
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (tag_id, lang)
);

-- ============================================
-- НОВЕЛЛЫ
-- ============================================

CREATE TABLE novels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    cover_image_key VARCHAR(255),
    translation_status translation_status NOT NULL DEFAULT 'ongoing',
    original_chapters_count INTEGER DEFAULT 0,
    release_year INTEGER,
    author VARCHAR(255),
    views_total BIGINT NOT NULL DEFAULT 0,
    views_daily INTEGER NOT NULL DEFAULT 0,
    rating_sum INTEGER NOT NULL DEFAULT 0,
    rating_count INTEGER NOT NULL DEFAULT 0,
    bookmarks_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE novel_localizations (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    alt_titles TEXT[], -- Массив альтернативных названий
    search_vector TSVECTOR, -- Для полнотекстового поиска
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (novel_id, lang)
);

-- Связи новелла-жанр и новелла-тег
CREATE TABLE novel_genres (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (novel_id, genre_id)
);

CREATE TABLE novel_tags (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (novel_id, tag_id)
);

-- ============================================
-- ГЛАВЫ
-- ============================================

CREATE TABLE chapters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    number DECIMAL(10, 2) NOT NULL, -- Поддержка дробных номеров (1.5, 2.1 и т.д.)
    slug VARCHAR(255),
    title VARCHAR(500),
    views INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (novel_id, number)
);

CREATE TABLE chapter_contents (
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    content TEXT NOT NULL,
    word_count INTEGER NOT NULL DEFAULT 0,
    source VARCHAR(50) NOT NULL DEFAULT 'manual', -- manual, auto, import
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chapter_id, lang)
);

-- ============================================
-- ПРОГРЕСС ЧТЕНИЯ
-- ============================================

CREATE TABLE reading_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    chapter_id UUID NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    position INTEGER NOT NULL DEFAULT 0, -- Позиция в тексте (scroll position)
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, novel_id)
);

-- ============================================
-- ЗАКЛАДКИ
-- ============================================

CREATE TABLE bookmark_lists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code bookmark_list_code NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, code)
);

CREATE TABLE bookmarks (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    list_id UUID NOT NULL REFERENCES bookmark_lists(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, novel_id)
);

-- ============================================
-- РЕЙТИНГИ
-- ============================================

CREATE TABLE novel_ratings (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value INTEGER NOT NULL CHECK (value >= 1 AND value <= 10),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (novel_id, user_id)
);

-- ============================================
-- ПРОСМОТРЫ (для трендов)
-- ============================================

CREATE TABLE novel_views_daily (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    views INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (novel_id, date)
);

-- ============================================
-- ИНДЕКСЫ
-- ============================================

-- Пользователи
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Новеллы
CREATE INDEX idx_novels_slug ON novels(slug);
CREATE INDEX idx_novels_translation_status ON novels(translation_status);
CREATE INDEX idx_novels_created_at ON novels(created_at DESC);
CREATE INDEX idx_novels_updated_at ON novels(updated_at DESC);
CREATE INDEX idx_novels_views_daily ON novels(views_daily DESC);
CREATE INDEX idx_novels_views_total ON novels(views_total DESC);
CREATE INDEX idx_novels_rating ON novels((rating_sum::float / NULLIF(rating_count, 0)) DESC NULLS LAST);

-- Локализации новелл - полнотекстовый поиск
CREATE INDEX idx_novel_localizations_search ON novel_localizations USING GIN(search_vector);
CREATE INDEX idx_novel_localizations_title ON novel_localizations USING GIN(title gin_trgm_ops);

-- Главы
CREATE INDEX idx_chapters_novel_id ON chapters(novel_id);
CREATE INDEX idx_chapters_number ON chapters(novel_id, number);
CREATE INDEX idx_chapters_published_at ON chapters(published_at DESC);

-- Прогресс чтения
CREATE INDEX idx_reading_progress_user_id ON reading_progress(user_id);
CREATE INDEX idx_reading_progress_novel_id ON reading_progress(novel_id);

-- Закладки
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_list_id ON bookmarks(list_id);
CREATE INDEX idx_bookmarks_updated_at ON bookmarks(updated_at DESC);

-- Просмотры
CREATE INDEX idx_novel_views_daily_date ON novel_views_daily(date DESC);

-- ============================================
-- ТРИГГЕРЫ
-- ============================================

-- Триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novels_updated_at
    BEFORE UPDATE ON novels
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_novel_localizations_updated_at
    BEFORE UPDATE ON novel_localizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chapters_updated_at
    BEFORE UPDATE ON chapters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chapter_contents_updated_at
    BEFORE UPDATE ON chapter_contents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bookmarks_updated_at
    BEFORE UPDATE ON bookmarks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для обновления search_vector в novel_localizations
CREATE OR REPLACE FUNCTION update_novel_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = 
        setweight(to_tsvector('simple', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(array_to_string(NEW.alt_titles, ' '), '')), 'A');
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_novel_localizations_search_vector
    BEFORE INSERT OR UPDATE ON novel_localizations
    FOR EACH ROW
    EXECUTE FUNCTION update_novel_search_vector();

-- Триггер для подсчета слов в главе
CREATE OR REPLACE FUNCTION update_chapter_word_count()
RETURNS TRIGGER AS $$
BEGIN
    NEW.word_count = array_length(regexp_split_to_array(NEW.content, '\s+'), 1);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_chapter_contents_word_count
    BEFORE INSERT OR UPDATE ON chapter_contents
    FOR EACH ROW
    EXECUTE FUNCTION update_chapter_word_count();

-- Триггер для обновления bookmarks_count в novels
CREATE OR REPLACE FUNCTION update_novel_bookmarks_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE novels SET bookmarks_count = bookmarks_count + 1 WHERE id = NEW.novel_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE novels SET bookmarks_count = bookmarks_count - 1 WHERE id = OLD.novel_id;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_novels_bookmarks_count
    AFTER INSERT OR DELETE ON bookmarks
    FOR EACH ROW
    EXECUTE FUNCTION update_novel_bookmarks_count();

-- Триггер для обновления рейтинга в novels
CREATE OR REPLACE FUNCTION update_novel_rating()
RETURNS TRIGGER AS $$
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
$$ language 'plpgsql';

CREATE TRIGGER update_novels_rating
    AFTER INSERT OR UPDATE OR DELETE ON novel_ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_novel_rating();
