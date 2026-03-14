-- 001_create_ledger.down.sql
DROP TRIGGER IF EXISTS double_entry_trigger ON journal_lines;
DROP FUNCTION IF EXISTS check_double_entry();
DROP TABLE IF EXISTS journal_lines;
DROP TABLE IF EXISTS journals;
DROP TABLE IF EXISTS accounts;
