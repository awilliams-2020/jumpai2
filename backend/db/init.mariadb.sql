-- Create database
CREATE DATABASE IF NOT EXISTS advisor_scheduling;

-- Use the database
USE advisor_scheduling;

-- Create users table
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    email VARCHAR(255) NOT NULL,
    google_id VARCHAR(255) UNIQUE,
    hubspot_id VARCHAR(255) UNIQUE,
    access_token TEXT,
    refresh_token TEXT,
    token_expiry TIMESTAMP NULL DEFAULT NULL,
    calendar_ids JSON,
    last_login_at TIMESTAMP NULL DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    UNIQUE KEY unique_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create scheduling_windows table
CREATE TABLE scheduling_windows (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    start_hour TINYINT UNSIGNED NOT NULL,
    end_hour TINYINT UNSIGNED NOT NULL,
    weekday TINYINT UNSIGNED NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT fk_scheduling_windows_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT valid_weekday CHECK (weekday <= 6),
    CONSTRAINT valid_hours CHECK (start_hour < end_hour AND start_hour <= 23 AND end_hour <= 23)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create scheduling_links table
CREATE TABLE scheduling_links (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(255) NOT NULL,
    duration SMALLINT UNSIGNED NOT NULL,
    max_uses INT UNSIGNED NULL DEFAULT NULL,
    expires_at TIMESTAMP NULL DEFAULT NULL,
    max_days_in_advance SMALLINT UNSIGNED NOT NULL,
    custom_questions JSON,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT fk_scheduling_links_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT positive_duration CHECK (duration > 0),
    CONSTRAINT positive_max_uses CHECK (max_uses IS NULL OR max_uses > 0),
    CONSTRAINT positive_max_days CHECK (max_days_in_advance > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create meetings table
CREATE TABLE meetings (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    scheduling_link_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,
    client_email VARCHAR(255) NOT NULL,
    linkedin_url VARCHAR(255) NULL DEFAULT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    answers JSON,
    hubspot_contact_id VARCHAR(255) NULL DEFAULT NULL,
    linkedin_data JSON,
    context_notes TEXT,
    CONSTRAINT fk_meetings_scheduling_link
        FOREIGN KEY (scheduling_link_id) REFERENCES scheduling_links(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_meetings_user
        FOREIGN KEY (user_id) REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT valid_time_range CHECK (start_time < end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_hubspot_id ON users(hubspot_id);
CREATE INDEX idx_scheduling_windows_user_id ON scheduling_windows(user_id);
CREATE INDEX idx_scheduling_links_user_id ON scheduling_links(user_id);
CREATE INDEX idx_meetings_scheduling_link_id ON meetings(scheduling_link_id);
CREATE INDEX idx_meetings_user_id ON meetings(user_id);
CREATE INDEX idx_meetings_start_time ON meetings(start_time);

-- Create stored procedure for soft delete
DELIMITER //
CREATE PROCEDURE soft_delete_user(IN user_id BIGINT)
BEGIN
    UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = user_id;
END //
DELIMITER ;

-- Create stored procedure for soft delete scheduling link
DELIMITER //
CREATE PROCEDURE soft_delete_scheduling_link(IN link_id BIGINT)
BEGIN
    UPDATE scheduling_links SET deleted_at = CURRENT_TIMESTAMP WHERE id = link_id;
END //
DELIMITER ;

-- Create stored procedure for soft delete meeting
DELIMITER //
CREATE PROCEDURE soft_delete_meeting(IN meeting_id BIGINT)
BEGIN
    UPDATE meetings SET deleted_at = CURRENT_TIMESTAMP WHERE id = meeting_id;
END //
DELIMITER ;

-- Create view for active scheduling windows
CREATE VIEW active_scheduling_windows AS
SELECT * FROM scheduling_windows
WHERE is_active = TRUE AND deleted_at IS NULL;

-- Create view for active scheduling links
CREATE VIEW active_scheduling_links AS
SELECT * FROM scheduling_links
WHERE is_active = TRUE 
AND deleted_at IS NULL
AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
AND (max_uses IS NULL OR max_uses > (
    SELECT COUNT(*) FROM meetings 
    WHERE meetings.scheduling_link_id = scheduling_links.id
    AND meetings.deleted_at IS NULL
)); 