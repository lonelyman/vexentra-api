-- scripts/seed_showcase.sql
INSERT INTO users (id, person_id, username, email, password_hash, created_at, updated_at) 
VALUES (gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'somchai', 'somchai@vexentra.co', 'hash', NOW(), NOW()) ON CONFLICT DO NOTHING;

INSERT INTO profiles (person_id, display_name, headline, bio, location, avatar_url, created_at, updated_at) 
VALUES ('019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'สมชาย วงษ์สวรรค์', 'Brand Designer & Creative Strategist', 'นักออกแบบแบรนด์ที่มีประสบการณ์กว่า 8 ปีในการสร้างอัตลักษณ์องค์กรให้กับธุรกิจ SME ถึงองค์กรขนาดใหญ่ เชื่อว่าการออกแบบที่ดีต้องสื่อเรื่องราว ไม่ใช่แค่ความสวยงาม', 'กรุงเทพฯ, ประเทศไทย', '🎨', NOW(), NOW()) ON CONFLICT DO NOTHING;

-- Skills
INSERT INTO skills (id, person_id, name, category, proficiency, sort_order) VALUES
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Brand Identity', 'การออกแบบ', 5, 1),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Visual Design', 'การออกแบบ', 4, 2),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Adobe Illustrator', 'เครื่องมือ', 5, 3),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Figma', 'เครื่องมือ', 4, 4),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Brand Strategy', 'กลยุทธ์และการสื่อสาร', 5, 5),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Client Management', 'ทักษะธุรกิจ', 5, 6)
ON CONFLICT DO NOTHING;

-- Experiences
INSERT INTO experiences (id, person_id, company, position, location, description, started_at, is_current, sort_order) VALUES
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Vexentra Studio', 'Founder & Creative Director', 'กรุงเทพฯ, ไทย', 'ก่อตั้งสตูดิโอออกแบบบแรนด์ของตัวเอง ดูแลงานตั้งแต่ Discovery Session ไปจนถึง Delivery', '2021-01-01', true, 1),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'BBDO Bangkok', 'Senior Brand Designer', 'กรุงเทพฯ, ไทย', 'ออกแบบและดูแลทิศทางสร้างสรรค์ให้กับแคมเปญระดับประเทศ', '2018-01-01', false, 2)
ON CONFLICT DO NOTHING;

-- Portfolio
INSERT INTO portfolio_items (id, person_id, title, slug, summary, cover_image_url, status, featured) VALUES
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Refresh Skincare — Complete Brand Identity', 'refresh-skincare', 'ออกแบบ Brand Identity ครบวงจรสำหรับแบรนด์ Skincare ธรรมชาติ', '🧴', 'published', true),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Greenway Properties — Corporate Rebranding', 'greenway-properties', 'Rebranding บริษัทอสังหาริมทรัพย์ขนาดกลางให้ทันสมัย', '🏢', 'published', false),
(gen_random_uuid(), '019d71bd-1c2d-7a2c-a1d7-fd84416611f9', 'Sabthai — F&B Brand Launch', 'sabthai', 'สร้างแบรนด์ร้านอาหารไทยใหม่ตั้งแต่ศูนย์', '🍜', 'published', true)
ON CONFLICT DO NOTHING;
