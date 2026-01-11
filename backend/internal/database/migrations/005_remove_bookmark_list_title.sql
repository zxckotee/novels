-- Migration: 005_remove_bookmark_list_title
-- Description: Remove title, is_system and updated_at columns from bookmark_lists table as they're now handled via localization and code-based logic
-- Also add id column to bookmarks table

-- Remove title column from bookmark_lists
ALTER TABLE bookmark_lists DROP COLUMN IF EXISTS title;

-- Remove is_system column from bookmark_lists
ALTER TABLE bookmark_lists DROP COLUMN IF EXISTS is_system;

-- Remove updated_at column from bookmark_lists
ALTER TABLE bookmark_lists DROP COLUMN IF EXISTS updated_at;

-- Add id column to bookmarks table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='bookmarks' AND column_name='id') THEN
        ALTER TABLE bookmarks ADD COLUMN id UUID DEFAULT gen_random_uuid();
    END IF;
END $$;
