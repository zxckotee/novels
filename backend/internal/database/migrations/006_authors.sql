-- ============================================
-- АВТОРЫ
-- ============================================

-- Таблица авторов
CREATE TABLE authors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Локализации авторов
CREATE TABLE author_localizations (
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    lang VARCHAR(10) NOT NULL,
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (author_id, lang)
);

-- Связь многие-ко-многим: новеллы и авторы
CREATE TABLE novel_authors (
    novel_id UUID NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (novel_id, author_id)
);

-- ============================================
-- ИНДЕКСЫ
-- ============================================

CREATE INDEX idx_authors_slug ON authors(slug);
CREATE INDEX idx_author_localizations_name ON author_localizations(name);
CREATE INDEX idx_novel_authors_novel_id ON novel_authors(novel_id);
CREATE INDEX idx_novel_authors_author_id ON novel_authors(author_id);
CREATE INDEX idx_novel_authors_is_primary ON novel_authors(is_primary);

-- ============================================
-- ТРИГГЕРЫ
-- ============================================

-- Триггер для обновления updated_at на authors
CREATE TRIGGER update_authors_updated_at
    BEFORE UPDATE ON authors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для обновления updated_at на author_localizations
CREATE TRIGGER update_author_localizations_updated_at
    BEFORE UPDATE ON author_localizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- МИГРАЦИЯ ДАННЫХ
-- ============================================

-- Мигрируем существующие данные из поля novels.author в таблицу authors
-- Создаем авторов из уникальных значений novels.author (если они есть)
DO $$
DECLARE
    v_author_name TEXT;
    v_author_slug TEXT;
    v_author_id UUID;
    v_novel_id UUID;
    v_novel_author TEXT;
BEGIN
    -- Проходим по всем новеллам с указанным автором
    FOR v_novel_id, v_novel_author IN 
        SELECT id, author FROM novels WHERE author IS NOT NULL AND author != ''
    LOOP
        v_author_name := v_novel_author;
        
        -- Генерируем slug для автора
        v_author_slug := lower(regexp_replace(v_author_name, '[^a-zA-Z0-9\s-]', '', 'g'));
        v_author_slug := regexp_replace(v_author_slug, '\s+', '-', 'g');
        v_author_slug := regexp_replace(v_author_slug, '-+', '-', 'g');
        v_author_slug := trim(both '-' from v_author_slug);
        
        -- Если slug пустой, используем UUID
        IF v_author_slug = '' THEN
            v_author_slug := 'author-' || substring(uuid_generate_v4()::text, 1, 8);
        END IF;
        
        -- Ищем существующего автора по slug
        SELECT id INTO v_author_id FROM authors WHERE slug = v_author_slug;
        
        -- Если автор не найден, создаем его
        IF v_author_id IS NULL THEN
            INSERT INTO authors (slug) VALUES (v_author_slug) RETURNING id INTO v_author_id;
            
            -- Добавляем локализацию для автора (русский по умолчанию)
            INSERT INTO author_localizations (author_id, lang, name)
            VALUES (v_author_id, 'ru', v_author_name);
        END IF;
        
        -- Создаем связь новелла-автор (если еще не создана)
        INSERT INTO novel_authors (novel_id, author_id, is_primary)
        VALUES (v_novel_id, v_author_id, TRUE)
        ON CONFLICT (novel_id, author_id) DO NOTHING;
    END LOOP;
END $$;

-- Комментарий: Поле novels.author можно будет удалить в будущей миграции
-- после проверки корректности данных в novel_authors
COMMENT ON COLUMN novels.author IS 'DEPRECATED: Use novel_authors table instead. Will be removed in future migration.';
