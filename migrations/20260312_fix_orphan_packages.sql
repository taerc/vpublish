-- Migration: Fix Orphan Packages
-- Description: Identify and fix packages that reference soft-deleted or non-existent categories
-- Date: 2026-03-12
-- Issue: Package id=9 has category_id=10 but category 10 is soft-deleted, causing missing category object in API response

-- ============================================
-- STEP 1: Identify orphan packages
-- ============================================
-- This query finds packages whose category is either:
-- 1. Non-existent (category_id points to nothing)
-- 2. Soft-deleted (category has deleted_at set)

SELECT 
    p.id as package_id,
    p.category_id,
    p.name as package_name,
    c.id as category_exists,
    c.name as category_name,
    c.deleted_at as category_deleted_at,
    CASE 
        WHEN c.id IS NULL THEN 'CATEGORY_NOT_FOUND'
        WHEN c.deleted_at IS NOT NULL THEN 'CATEGORY_SOFT_DELETED'
        ELSE 'OK'
    END as status
FROM packages p
LEFT JOIN categories c ON p.category_id = c.id
WHERE c.id IS NULL OR c.deleted_at IS NOT NULL;

-- ============================================
-- STEP 2: Fix options (choose one)
-- ============================================

-- Option A: Reassign orphan packages to a new category
-- First, create a new category for orphan packages if it doesn't exist
-- INSERT INTO categories (name, code, description, sort_order, is_active, created_at, updated_at)
-- VALUES ('未分类', 'TYPE_WEI_FEN_LEI', '孤立软件包分类', 999, 1, NOW(), NOW());

-- Then update orphan packages to use the new category
-- UPDATE packages p
-- LEFT JOIN categories c ON p.category_id = c.id
-- SET p.category_id = (SELECT id FROM categories WHERE code = 'TYPE_WEI_FEN_LEI')
-- WHERE c.id IS NULL OR c.deleted_at IS NOT NULL;

-- Option B: Delete orphan packages (USE WITH CAUTION)
-- DELETE p FROM packages p
-- LEFT JOIN categories c ON p.category_id = c.id
-- WHERE c.id IS NULL OR c.deleted_at IS NOT NULL;

-- Option C: Restore soft-deleted categories
-- This restores the category but keeps the deletion history
-- UPDATE categories SET deleted_at = NULL WHERE id = 10;

-- ============================================
-- STEP 3: Recommended Fix for Current Issue
-- ============================================
-- For the specific case of category_id=10 being soft-deleted:
-- Restore the category to allow proper package management

-- Uncomment the following to restore category 10:
-- UPDATE categories SET deleted_at = NULL WHERE id = 10;

-- ============================================
-- STEP 4: Verification Query
-- ============================================
-- Run this after applying fixes to verify no orphan packages remain

SELECT 
    COUNT(*) as orphan_count,
    'Should be 0 after fix' as expected
FROM packages p
LEFT JOIN categories c ON p.category_id = c.id
WHERE c.id IS NULL OR c.deleted_at IS NOT NULL;
