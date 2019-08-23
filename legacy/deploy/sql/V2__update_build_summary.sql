ALTER TABLE build_summary ADD COLUMN status INT;
UPDATE build_summary AS bs
  SET status = (CASE
                  WHEN bs.failed = false
                  THEN 4
                  ELSE 3
                END)
;