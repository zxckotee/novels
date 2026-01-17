-- ============================================
-- НАСТРОЙКИ ПРИЛОЖЕНИЯ И АУДИТ
-- ============================================

-- Таблица настроек приложения
CREATE TABLE app_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Таблица логов действий администраторов
CREATE TABLE admin_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- ============================================
-- ИНДЕКСЫ
-- ============================================

CREATE INDEX idx_admin_audit_log_actor ON admin_audit_log(actor_user_id);
CREATE INDEX idx_admin_audit_log_action ON admin_audit_log(action);
CREATE INDEX idx_admin_audit_log_entity ON admin_audit_log(entity_type, entity_id);
CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log(created_at DESC);

-- ============================================
-- ТРИГГЕРЫ
-- ============================================

-- Триггер для обновления updated_at в app_settings
CREATE TRIGGER update_app_settings_updated_at
    BEFORE UPDATE ON app_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- ДЕФОЛТНЫЕ НАСТРОЙКИ
-- ============================================

-- Вставляем базовые настройки приложения
INSERT INTO app_settings (key, value, description) VALUES
    ('site_name', '{"ru": "Новеллы", "en": "Novels"}'::jsonb, 'Название сайта'),
    ('site_description', '{"ru": "Платформа для чтения новелл", "en": "Platform for reading novels"}'::jsonb, 'Описание сайта'),
    ('registration_enabled', 'true'::jsonb, 'Разрешена ли регистрация новых пользователей'),
    ('comments_enabled', 'true'::jsonb, 'Включены ли комментарии на сайте'),
    ('maintenance_mode', 'false'::jsonb, 'Режим технического обслуживания'),
    ('max_upload_size_mb', '10'::jsonb, 'Максимальный размер загружаемого файла в МБ'),
    ('chapters_per_day_limit', '0'::jsonb, 'Лимит глав в день для обычных пользователей (0 = без лимита)'),
    ('xp_multiplier', '1.0'::jsonb, 'Множитель для начисления XP'),
    ('featured_novels_count', '6'::jsonb, 'Количество рекомендуемых новелл на главной'),
    ('trending_days_window', '7'::jsonb, 'Окно в днях для расчета трендовых новелл'),
    ('comments_per_page', '20'::jsonb, 'Количество комментариев на странице'),
    ('novels_per_page', '24'::jsonb, 'Количество новелл на странице каталога'),
    ('min_comment_length', '1'::jsonb, 'Минимальная длина комментария'),
    ('max_comment_length', '10000'::jsonb, 'Максимальная длина комментария'),
    ('email_verification_required', 'false'::jsonb, 'Требуется ли подтверждение email при регистрации'),
    ('analytics_enabled', 'false'::jsonb, 'Включена ли аналитика'),
    ('cdn_url', 'null'::jsonb, 'URL CDN для статических файлов'),
    ('backup_retention_days', '30'::jsonb, 'Количество дней хранения резервных копий')
ON CONFLICT (key) DO NOTHING;

-- ============================================
-- ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
-- ============================================

-- Функция для логирования админ действий
CREATE OR REPLACE FUNCTION log_admin_action(
    p_actor_user_id UUID,
    p_action VARCHAR,
    p_entity_type VARCHAR,
    p_entity_id UUID DEFAULT NULL,
    p_details JSONB DEFAULT NULL,
    p_ip_address VARCHAR DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
)
RETURNS UUID AS $$
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
$$ LANGUAGE plpgsql;

-- Функция для получения настройки
CREATE OR REPLACE FUNCTION get_setting(p_key VARCHAR)
RETURNS JSONB AS $$
DECLARE
    v_value JSONB;
BEGIN
    SELECT value INTO v_value FROM app_settings WHERE key = p_key;
    RETURN v_value;
END;
$$ LANGUAGE plpgsql;

-- Функция для обновления настройки
CREATE OR REPLACE FUNCTION update_setting(
    p_key VARCHAR,
    p_value JSONB,
    p_updated_by UUID DEFAULT NULL
)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE app_settings 
    SET value = p_value, updated_by = p_updated_by, updated_at = NOW()
    WHERE key = p_key;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;
