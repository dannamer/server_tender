-- Active: 1726220107504@@192.168.0.11@5432@your_database
-- SELECT EXISTS (
-- 			SELECT 1 
-- 			FROM organization_responsible 
-- 			WHERE user_id = $1
-- 		)
SELECT
  e.id,
  e.username,
  e.first_name,
  e.last_name,
  e.created_at,
  e.updated_at,
  eo.organization_id
FROM
  employee e
  LEFT JOIN organization_responsible  eo ON e.id = eo.user_id
WHERE
  e.id = "550e8400-e29b-41d4-a716-446655440001"