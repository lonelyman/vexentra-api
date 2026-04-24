-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS files (
  id                 uuid PRIMARY KEY,
  owner_type         text NOT NULL,
  owner_id           uuid NOT NULL,
  category           text NOT NULL,
  object_key         text NOT NULL UNIQUE,
  original_filename  text NOT NULL,
  mime_type          text NOT NULL,
  size_bytes         bigint NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 31457280),
  sha256             text NOT NULL,
  etag               text,
  visibility         text NOT NULL DEFAULT 'private' CHECK (visibility IN ('private')),
  processing_status  text NOT NULL DEFAULT 'ready' CHECK (processing_status IN ('pending', 'ready', 'failed')),
  processing_error   text,
  metadata           jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_by         uuid NOT NULL REFERENCES users(id),
  created_at         timestamp with time zone NOT NULL DEFAULT now(),
  updated_at         timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at         timestamp with time zone
);

CREATE INDEX IF NOT EXISTS idx_files_owner_category
  ON files (owner_type, owner_id, category)
  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files (deleted_at);

CREATE TABLE IF NOT EXISTS upload_sessions (
  id                 uuid PRIMARY KEY,
  user_id            uuid NOT NULL REFERENCES users(id),
  person_id          uuid NOT NULL REFERENCES persons(id),
  intent             text NOT NULL CHECK (intent IN ('profile_image')),
  temp_object_key    text NOT NULL UNIQUE,
  original_filename  text NOT NULL,
  expected_mime      text NOT NULL,
  expected_max_size  bigint NOT NULL CHECK (expected_max_size > 0 AND expected_max_size <= 31457280),
  status             text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'expired', 'cancelled')),
  expires_at         timestamp with time zone NOT NULL,
  completed_at       timestamp with time zone,
  created_at         timestamp with time zone NOT NULL DEFAULT now(),
  updated_at         timestamp with time zone NOT NULL DEFAULT now(),
  deleted_at         timestamp with time zone
);

CREATE INDEX IF NOT EXISTS idx_upload_sessions_user_status
  ON upload_sessions (user_id, status)
  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_upload_sessions_person_intent
  ON upload_sessions (person_id, intent)
  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_upload_sessions_expires_at
  ON upload_sessions (expires_at);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_deleted_at
  ON upload_sessions (deleted_at);

ALTER TABLE profiles
  ADD COLUMN IF NOT EXISTS avatar_file_id uuid REFERENCES files(id) ON DELETE SET NULL;

ALTER TABLE profiles
  DROP COLUMN IF EXISTS avatar_url;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE profiles
  ADD COLUMN IF NOT EXISTS avatar_url text;

ALTER TABLE profiles
  DROP COLUMN IF EXISTS avatar_file_id;

DROP INDEX IF EXISTS idx_upload_sessions_deleted_at;
DROP INDEX IF EXISTS idx_upload_sessions_expires_at;
DROP INDEX IF EXISTS idx_upload_sessions_person_intent;
DROP INDEX IF EXISTS idx_upload_sessions_user_status;
DROP TABLE IF EXISTS upload_sessions;

DROP INDEX IF EXISTS idx_files_deleted_at;
DROP INDEX IF EXISTS idx_files_owner_category;
DROP TABLE IF EXISTS files;
-- +goose StatementEnd
