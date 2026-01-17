-- ============================================
-- УНИФИКАЦИЯ СХЕМЫ КОММЕНТАРИЕВ
-- ============================================
-- Приводим таблицу comments к схеме, используемой в models.Comment

-- Добавляем новые поля для новой схемы (target_type, target_id, body, is_deleted, root_id, depth)
ALTER TABLE comments
    ADD COLUMN IF NOT EXISTS target_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS target_id UUID,
    ADD COLUMN IF NOT EXISTS body TEXT,
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS root_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS depth INTEGER NOT NULL DEFAULT 0;

-- Мигрируем данные из старых полей
UPDATE comments
SET 
    target_type = CASE 
        WHEN novel_id IS NOT NULL THEN 'novel'
        WHEN chapter_id IS NOT NULL THEN 'chapter'
    END,
    target_id = COALESCE(novel_id, chapter_id),
    body = content,
    is_deleted = (deleted_at IS NOT NULL)
WHERE target_type IS NULL;

-- Обновляем root_id и depth для существующих комментариев
-- Сначала для корневых комментариев (без parent_id)
UPDATE comments
SET root_id = id, depth = 0
WHERE parent_id IS NULL AND root_id IS NULL;

-- Затем для комментариев первого уровня вложенности
UPDATE comments c
SET root_id = c.parent_id, depth = 1
WHERE c.parent_id IS NOT NULL 
  AND c.root_id IS NULL
  AND EXISTS (SELECT 1 FROM comments p WHERE p.id = c.parent_id AND p.parent_id IS NULL);

-- Для более глубоких уровней - запускаем итеративно (до 10 уровней)
DO $$
DECLARE
    i INTEGER;
    updated_count INTEGER;
BEGIN
    FOR i IN 2..10 LOOP
        UPDATE comments c
        SET root_id = p.root_id, depth = i
        FROM comments p
        WHERE c.parent_id = p.id
          AND p.depth = i - 1
          AND c.root_id IS NULL;
        
        GET DIAGNOSTICS updated_count = ROW_COUNT;
        EXIT WHEN updated_count = 0;
    END LOOP;
END $$;

-- Делаем новые поля обязательными
ALTER TABLE comments
    ALTER COLUMN target_type SET NOT NULL,
    ALTER COLUMN target_id SET NOT NULL,
    ALTER COLUMN body SET NOT NULL;

-- Удаляем CHECK constraint на старых полях (если он есть)
ALTER TABLE comments
    DROP CONSTRAINT IF EXISTS comments_check;

-- Создаем новый индекс для target_type + target_id
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_comments_target_id ON comments(target_id);
CREATE INDEX IF NOT EXISTS idx_comments_root_id ON comments(root_id);
CREATE INDEX IF NOT EXISTS idx_comments_depth ON comments(depth);
CREATE INDEX IF NOT EXISTS idx_comments_is_deleted ON comments(is_deleted) WHERE is_deleted = TRUE;

-- Добавляем комментарии для старых полей (пометка устаревших)
COMMENT ON COLUMN comments.novel_id IS 'DEPRECATED: Use target_type=novel and target_id instead. Will be removed in future migration.';
COMMENT ON COLUMN comments.chapter_id IS 'DEPRECATED: Use target_type=chapter and target_id instead. Will be removed in future migration.';
COMMENT ON COLUMN comments.content IS 'DEPRECATED: Use body instead. Will be removed in future migration.';
COMMENT ON COLUMN comments.deleted_at IS 'DEPRECATED: Use is_deleted instead. Will be removed in future migration.';

-- ============================================
-- ОБНОВЛЕНИЕ ТРИГГЕРОВ
-- ============================================

-- Обновляем триггер для обновления replies_count с учетом is_deleted
DROP TRIGGER IF EXISTS trigger_comment_replies_count ON comments;

CREATE OR REPLACE FUNCTION update_comment_replies_count()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_comment_replies_count
    AFTER INSERT OR UPDATE OR DELETE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_comment_replies_count();

-- Trigg для автоматического установления root_id и depth при создании
CREATE OR REPLACE FUNCTION set_comment_hierarchy()
RETURNS TRIGGER AS $$
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
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_set_comment_hierarchy ON comments;
CREATE TRIGGER trigger_set_comment_hierarchy
    BEFORE INSERT ON comments
    FOR EACH ROW EXECUTE FUNCTION set_comment_hierarchy();
