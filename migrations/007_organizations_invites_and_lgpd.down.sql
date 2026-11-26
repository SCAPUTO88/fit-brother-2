-- migrations/007_organizations_invites_and_lgpd.down.sql

DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS privacy_requests;
DROP TABLE IF EXISTS consent_records;
DROP TABLE IF EXISTS user_mfa;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS trainer_student_relationships;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS organizations;
