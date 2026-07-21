CREATE TABLE IF NOT EXISTS repositories (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  origin VARCHAR(512) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS commit_records (
  repository_id BIGINT UNSIGNED NOT NULL,
  commit_id VARCHAR(64) NOT NULL,
  author VARCHAR(255) NOT NULL,
  ai_lines INT UNSIGNED NOT NULL,
  total_lines INT UNSIGNED NOT NULL,
  is_ai_commit BOOLEAN NOT NULL,
  committed_at DATETIME NOT NULL,
  message TEXT NOT NULL,
  received_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (repository_id, commit_id),
  INDEX idx_commit_records_repository_date (repository_id, committed_at),
  INDEX idx_commit_records_author_date (author, committed_at),
  CONSTRAINT fk_commit_records_repository
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
