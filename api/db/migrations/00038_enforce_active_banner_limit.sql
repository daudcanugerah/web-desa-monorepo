-- +goose Up
-- +goose StatementBegin
-- Enforce max 20 active banners via trigger instead of application-level check
-- (avoids race conditions between count and update).
CREATE OR REPLACE FUNCTION enforce_active_banner_limit()
RETURNS TRIGGER AS $$
DECLARE
    active_count INTEGER;
BEGIN
    IF NEW.status = 'active' THEN
        SELECT COUNT(*) INTO active_count
        FROM banners
        WHERE status = 'active' AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::uuid);

        IF active_count >= 20 THEN
            RAISE EXCEPTION 'cannot activate banner: maximum of 20 active banners reached'
                USING ERRCODE = 'check_violation';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_active_banner_limit ON banners;
CREATE TRIGGER trg_active_banner_limit
BEFORE INSERT OR UPDATE OF status ON banners
FOR EACH ROW
WHEN (NEW.status = 'active')
EXECUTE FUNCTION enforce_active_banner_limit();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_active_banner_limit ON banners;
DROP FUNCTION IF EXISTS enforce_active_banner_limit();
-- +goose StatementEnd