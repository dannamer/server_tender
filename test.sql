-- Active: 1726220107504@@192.168.0.11@5432@your_database
SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible 
			WHERE user_id = $1
		)