--
-- PostgreSQL database dump
--

\restrict hFDahcxSxpDfZoCzdr4dRZeCD1PY303I2XnG5QeOUbwHxsX6SJG3seYFAkyZwhN

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: experiences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.experiences (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    company text NOT NULL,
    "position" text NOT NULL,
    location character varying(100),
    description text,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    is_current boolean DEFAULT false,
    sort_order bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: persons; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.persons (
    id uuid NOT NULL,
    name character varying(200) NOT NULL,
    invite_email character varying(254),
    invite_token character varying(128),
    invite_token_expires_at timestamp with time zone,
    linked_user_id uuid,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: portfolio_item_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.portfolio_item_tags (
    portfolio_item_id uuid NOT NULL,
    tag_id uuid NOT NULL
);


--
-- Name: portfolio_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.portfolio_items (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    title text NOT NULL,
    slug text NOT NULL,
    summary text,
    description text,
    content_markdown text,
    cover_image_url text,
    demo_url text,
    source_url text,
    status text DEFAULT 'draft'::text,
    featured boolean DEFAULT false,
    sort_order bigint DEFAULT 0,
    started_at timestamp with time zone,
    ended_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: portfolio_tags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.portfolio_tags (
    id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL
);


--
-- Name: profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.profiles (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    display_name character varying(100),
    headline text,
    bio text,
    location text,
    avatar_url text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: skills; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.skills (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    name text NOT NULL,
    category text DEFAULT 'other'::text,
    proficiency bigint DEFAULT 1,
    sort_order bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: social_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.social_links (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    platform_id uuid NOT NULL,
    url character varying(512) NOT NULL,
    sort_order bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: social_platforms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.social_platforms (
    id uuid NOT NULL,
    key text NOT NULL,
    name text NOT NULL,
    icon_url text,
    sort_order bigint DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: user_auths; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_auths (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider character varying(30) NOT NULL,
    provider_id character varying(254),
    secret character varying(60),
    refresh_token character varying(512)
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    person_id uuid NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(254) NOT NULL,
    status character varying(30) DEFAULT 'pending_verification'::character varying NOT NULL,
    last_login_at timestamp with time zone,
    is_email_verified boolean DEFAULT false NOT NULL,
    email_verification_token character varying(255),
    email_verification_token_expires_at timestamp with time zone,
    password_reset_token character varying(255),
    password_reset_token_expires_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    role character varying(20) DEFAULT 'user'::character varying NOT NULL
);


--
-- Data for Name: experiences; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.experiences (id, person_id, company, "position", location, description, started_at, ended_at, is_current, sort_order, created_at, updated_at) FROM stdin;
e83293d7-8bb7-48c8-a21d-5d7f602a1b65	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	บริษัท นูทริชั่น โปรเฟส จำกัด (มหาชน)	Fullstack Developer Chief	กรุงเทพฯ, ไทย	รับผิดชอบการออกแบบสถาปัตยกรรมระบบ (System Architecture) นำทีมพัฒนาระบบ Full-Stack ตลอดจนดูแลประสิทธิภาพการทำงานของเซิร์ฟเวอร์และฐานข้อมูลให้รองรับการเติบโตของธุรกิจ	2024-01-01 00:00:00+00	\N	t	1	\N	\N
dc8b0a68-c0f7-4b07-92ce-99755ae0d85c	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	บริษัท เจนเนติกพลัส จำกัด	Senior Programmer	กรุงเทพฯ, ไทย	พัฒนาระบบและแก้ไขปัญหาที่ซับซ้อน (Complex Problems) เป็นหลัก ให้คำปรึกษาแก่ทีมพัฒนา รวมถึงปรับปรุงระบบและฐานข้อมูลให้ทำงานได้อย่างรวดเร็ว มีเสถียรภาพ และปลอดภัยมากยิ่งขึ้น	2019-01-01 00:00:00+00	2024-01-01 00:00:00+00	f	2	\N	\N
448a1719-2dc8-4635-ad8d-96bd9662daee	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	บริษัท ฟิกท์ แอสโซซิเอท จำกัด	Senior Programmer	กรุงเทพฯ, ไทย	พัฒนาและดูแลรักษาระบบแอปพลิเคชันให้สอดคล้องกับความต้องการของธุรกิจ ร่วมออกแบบโครงสร้างฐานข้อมูลและ API เพื่อความยืดหยุ่นต่อการเพิ่มฟีเจอร์ใหม่	2016-01-01 00:00:00+00	2019-01-01 00:00:00+00	f	3	\N	\N
2d4d79c1-d8ec-4bd8-a0a0-7be5fd2939f9	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	บริษัท โมโนซอฟต์ จำกัด	Programmer	กรุงเทพฯ, ไทย	พัฒนาแอปพลิเคชันตามความต้องการของลูกค้าและองค์กร เขียนโค้ดส่วนการทำงานหลัก ตลอดจนทดสอบและแก้ไขข้อผิดพลาด (Bug Fixing) เพื่อให้ระบบพร้อมสำหรับการใช้งานจริง	2014-01-01 00:00:00+00	2016-01-01 00:00:00+00	f	4	\N	\N
\.


--
-- Data for Name: persons; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.persons (id, name, invite_email, invite_token, invite_token_expires_at, linked_user_id, created_by_user_id, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: portfolio_item_tags; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.portfolio_item_tags (portfolio_item_id, tag_id) FROM stdin;
1cb3607c-2830-4597-a7a2-1e4823cabfe6	2dbe0d1f-486a-4347-8454-4207cbfc2319
1cb3607c-2830-4597-a7a2-1e4823cabfe6	1f0dbb6f-518a-4349-b211-fc44323f3b0e
1cb3607c-2830-4597-a7a2-1e4823cabfe6	3f969b42-51e3-44ef-b29e-365541a56b6c
2eadf305-c55d-47ae-9504-d0a3934cb36c	02e98682-6488-4df8-88bb-3628481bb090
2eadf305-c55d-47ae-9504-d0a3934cb36c	83a99434-4555-44e7-88ee-8a2ac6c6b678
2eadf305-c55d-47ae-9504-d0a3934cb36c	171f4cbe-77a2-4b5d-accf-1c9535d34eda
2eadf305-c55d-47ae-9504-d0a3934cb36c	23594871-bc1c-430d-828d-9f591c0ed5e8
2eadf305-c55d-47ae-9504-d0a3934cb36c	f2971c9e-5a4e-490e-9858-98365a6a0106
d6e30b9f-c15f-4e40-8786-0f536d211bd4	2dbe0d1f-486a-4347-8454-4207cbfc2319
d6e30b9f-c15f-4e40-8786-0f536d211bd4	1f0dbb6f-518a-4349-b211-fc44323f3b0e
d6e30b9f-c15f-4e40-8786-0f536d211bd4	03652ef4-f860-4c4d-a979-9fd63a200fa3
babcfab4-656b-4d27-9b92-5e6742c76607	2dbe0d1f-486a-4347-8454-4207cbfc2319
babcfab4-656b-4d27-9b92-5e6742c76607	03652ef4-f860-4c4d-a979-9fd63a200fa3
babcfab4-656b-4d27-9b92-5e6742c76607	073d86ec-6571-4858-8345-380830452482
dbd08713-0a4e-4c10-98b8-754274b98b2e	03652ef4-f860-4c4d-a979-9fd63a200fa3
dbd08713-0a4e-4c10-98b8-754274b98b2e	0ca1e791-4a66-4fff-ae7a-4832daf703c3
dbd08713-0a4e-4c10-98b8-754274b98b2e	f14f114f-4924-4153-a499-5e689ad77a8a
dbd08713-0a4e-4c10-98b8-754274b98b2e	dd896779-d4ca-4c5d-a1a9-0553b0e2f033
4a0714ee-578a-40b4-8868-c0b5a6371abd	2dbe0d1f-486a-4347-8454-4207cbfc2319
4a0714ee-578a-40b4-8868-c0b5a6371abd	03652ef4-f860-4c4d-a979-9fd63a200fa3
4a0714ee-578a-40b4-8868-c0b5a6371abd	073d86ec-6571-4858-8345-380830452482
e432a29f-3d7d-4ab7-9725-e7501fb456a0	83a99434-4555-44e7-88ee-8a2ac6c6b678
e432a29f-3d7d-4ab7-9725-e7501fb456a0	0ca1e791-4a66-4fff-ae7a-4832daf703c3
e432a29f-3d7d-4ab7-9725-e7501fb456a0	dfa19d67-635d-4b85-9cf9-5d147eaeedb1
146564d1-bbfc-4299-874a-a195f7dfa8d1	9f31b519-96dc-404c-a8a4-0584ce67c398
146564d1-bbfc-4299-874a-a195f7dfa8d1	f706d934-faa7-4364-b174-bf2119f0332d
\.


--
-- Data for Name: portfolio_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.portfolio_items (id, person_id, title, slug, summary, description, content_markdown, cover_image_url, demo_url, source_url, status, featured, sort_order, started_at, ended_at, created_at, updated_at, deleted_at) FROM stdin;
1cb3607c-2830-4597-a7a2-1e4823cabfe6	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	AdSpace Pro — Media Management System	adspace-pro-media-management	ระบบบริหารจัดการพื้นที่สื่อโฆษณา (Ad Space) ตรวจสอบสถานะการเช่า พื้นที่ว่าง และแจ้งเตือนสัญญาเช่า	พัฒนาระบบบริหารจัดการป้ายโฆษณาและพื้นที่สื่อ (Media Management System) เพื่อใช้สำหรับติดตามสถานะพื้นที่เช่า ว่าง/ถูกเช่า รวมถึงระบบแจ้งเตือนวันหมดอายุของสัญญา และจัดการรายได้ พัฒนาด้วย PHP CodeIgniter และ SQL Server	\N	📊	\N	\N	published	t	0	\N	\N	\N	\N	\N
2eadf305-c55d-47ae-9504-d0a3934cb36c	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	OMS — Order Management System	fxig08-order-management	ระบบบริหารจัดการคำสั่งซื้อ (Order Management) ครบวงจรสำหรับติดตามและจัดการออเดอร์	พัฒนาระบบบริหารจัดการคำสั่งซื้อ (Order Management System) เพื่อใช้สำหรับจัดการออเดอร์ ติดตามสถานะ และจัดการข้อมูลอย่างเป็นระบบ พัฒนาด้วย React และ Go Fiber	\N	📦	\N	\N	published	f	0	\N	\N	\N	\N	\N
d6e30b9f-c15f-4e40-8786-0f536d211bd4	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	TeamSync — Project Management System	teamsync-project-management	ระบบบริหารจัดการโครงการและติดตามความคืบหน้า (Project Tracker) สำหรับทีมพัฒนา	พัฒนาระบบสำหรับการกรอกข้อมูลความคืบหน้าของโครงการ ช่วยให้เห็นภาพรวมว่าใครรับผิดชอบงานส่วนไหน สถานะถึงไหนแล้ว รวมถึงเป็นช่องทางสื่อสารระหว่างคนในทีมเพื่อลดข้อผิดพลาด พัฒนาด้วย PHP CodeIgniter และ MySQL	\N	📈	\N	\N	published	t	0	\N	\N	\N	\N	\N
babcfab4-656b-4d27-9b92-5e6742c76607	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Charity Lottery Sales System	charity-lottery-sales	ระบบจำหน่ายสลากการกุศลออนไลน์ พร้อมระบบตรวจรางวัลอัตโนมัติ	พัฒนาระบบขายสลากการกุศล รองรับการซื้อขายผ่านช่องทางออนไลน์ มีระบบจัดการสลากหลังบ้าน และฟังก์ชันตรวจรางวัลอัตโนมัติ พัฒนาด้วย PHP CodeIgniter 3 และ MySQL	\N	🎟️	\N	\N	published	f	0	\N	\N	\N	\N	\N
dbd08713-0a4e-4c10-98b8-754274b98b2e	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Corporate PR & News Portal	corporate-pr-news-portal	ระบบประชาสัมพันธ์ข่าวสารและประกาศสำหรับองค์กรระดับประเทศ	พัฒนาระบบ Portal สำหรับจัดการและเผยแพร่ข่าวสารประชาสัมพันธ์ขององค์กรขนาดใหญ่ รองรับผู้ใช้งานจำนวนมาก โครงสร้างระบบรองรับ SEO และการแสดงผลที่รวดเร็ว พัฒนาด้วย Next.js, Node.js และ Sequelize ORM	\N	✈️	\N	\N	published	f	0	\N	\N	\N	\N	\N
4a0714ee-578a-40b4-8868-c0b5a6371abd	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Fleet & Machinery Maintenance System	fleet-machinery-maintenance	ระบบบริหารจัดการรอบซ่อมบำรุงยานพาหนะและเครื่องจักร	พัฒนาระบบซ่อมบำรุงสำหรับองค์กรขนาดใหญ่ที่มีรถยนต์และเครื่องจักรจำนวนมาก ช่วยคำนวณและแจ้งเตือนรอบการซ่อมบำรุง (Preventive Maintenance) จัดตารางนัดหมายช่างซ่อม และระบุตัวผู้รับผิดชอบในแต่ละแผนกได้อย่างชัดเจน พัฒนาด้วย PHP CodeIgniter 3 และ MySQL	\N	🚜	\N	\N	published	f	0	\N	\N	\N	\N	\N
e432a29f-3d7d-4ab7-9725-e7501fb456a0	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	OT & Allowance Management System	ot-allowance-management	ระบบบริหารจัดการเบิกจ่ายค่าล่วงเวลาและค่าตอบแทนพิเศษ	พัฒนาระบบสำหรับการจัดการคำร้องขอทำงานล่วงเวลา (OT) และการเบิกเงินพิเศษสำหรับองค์กรขนาดใหญ่ ช่วยคัดกรองและตรวจสอบเงื่อนไขสิทธิ์การเบิกจ่ายอัตโนมัติ เพื่อให้กระบวนการอนุมัติรวดเร็วและแม่นยำขึ้น พัฒนาด้วย Next.js และ Go (Gin Framework)	\N	⏱️	\N	\N	published	f	0	\N	\N	\N	\N	\N
146564d1-bbfc-4299-874a-a195f7dfa8d1	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Facebook Chat Analytics (POC)	facebook-chat-analytics-poc	โปรเจกต์ POC ดึงข้อมูลแชทผ่าน Facebook API เพื่อวิเคราะห์ข้อมูล	พัฒนาระบบ Proof of Concept (POC) สำหรับเชื่อมต่อกับ Facebook Graph API เพื่อดึงข้อมูลการสนทนามาจัดเก็บและวิเคราะห์ข้อมูลเบื้องต้น เช่น สถิติการพูดคุยว่าสนทนากับใครบ้าง	\N	💬	\N	\N	published	f	0	\N	\N	\N	\N	\N
\.


--
-- Data for Name: portfolio_tags; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.portfolio_tags (id, name, slug) FROM stdin;
2dbe0d1f-486a-4347-8454-4207cbfc2319	PHP	php
1f0dbb6f-518a-4349-b211-fc44323f3b0e	CodeIgniter	codeigniter
3f969b42-51e3-44ef-b29e-365541a56b6c	SQL Server	sql-server
02e98682-6488-4df8-88bb-3628481bb090	React	react
83a99434-4555-44e7-88ee-8a2ac6c6b678	Go	go
171f4cbe-77a2-4b5d-accf-1c9535d34eda	Fiber	fiber
23594871-bc1c-430d-828d-9f591c0ed5e8	GORM	gorm
f2971c9e-5a4e-490e-9858-98365a6a0106	PostgreSQL	postgresql
03652ef4-f860-4c4d-a979-9fd63a200fa3	MySQL	mysql
073d86ec-6571-4858-8345-380830452482	CodeIgniter 3	codeigniter3
0ca1e791-4a66-4fff-ae7a-4832daf703c3	Next.js	nextjs
f14f114f-4924-4153-a499-5e689ad77a8a	Node.js	nodejs
dd896779-d4ca-4c5d-a1a9-0553b0e2f033	Sequelize	sequelize
dfa19d67-635d-4b85-9cf9-5d147eaeedb1	Gin	gin
9f31b519-96dc-404c-a8a4-0584ce67c398	Facebook API	facebook-api
f706d934-faa7-4364-b174-bf2119f0332d	Data Analysis	data-analysis
\.


--
-- Data for Name: profiles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.profiles (id, person_id, display_name, headline, bio, location, avatar_url, created_at, updated_at) FROM stdin;
1a8a8a13-7034-4396-8013-c4f2ee07f3c4	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	นิพนธ์ คนสันเทียะ	Backend-Focused Full-Stack Developer	โปรแกรมเมอร์ผู้เชี่ยวชาญด้านการพัฒนาระบบ (Backend) และสถาปัตยกรรมซอฟต์แวร์ที่มั่นคง มีประสบการณ์ในการทำ Full-Stack Development โดยมุ่งเน้นไปที่ประสิทธิภาพ (Performance), ความปลอดภัย, และการออกแบบโครงสร้างที่รองรับการสเกลเป็นหลัก	กรุงเทพฯ, ประเทศไทย	💻	2026-04-20 03:46:06.194894+00	2026-04-20 03:46:06.194894+00
\.


--
-- Data for Name: skills; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.skills (id, person_id, name, category, proficiency, sort_order, created_at, updated_at) FROM stdin;
380b0373-402c-4286-9365-0f7e1c2a01d3	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Go (Golang)	Backend & API	5	1	\N	\N
2091f3d5-e61a-44fe-b6b7-8da971ad2fe6	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Fiber	Backend & API	5	2	\N	\N
3d3533e9-7a4b-4f74-8038-10b983f36345	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	GORM	Backend & API	5	3	\N	\N
8421d336-7ea9-4eae-ba76-b7c2267b6906	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Node.js	Backend & API	3	4	\N	\N
e73534f9-fea5-40d6-bb5d-ca1c99148335	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	PostgreSQL	Database	5	5	\N	\N
b033d91b-2573-4603-a953-98eb1448bf88	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	MySQL	Database	5	6	\N	\N
9dd3e894-2b57-4265-bd9d-e38edfef6805	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	SQL Server	Database	4	7	\N	\N
07e7f389-44b3-4322-8105-671c64c0321a	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Docker	DevOps & Tools	4	8	\N	\N
583bcf31-df0a-4ede-ad89-78262b727d10	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Git	DevOps & Tools	5	9	\N	\N
3e28aefa-206a-4ad5-8068-9887db586310	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Next.js / React	Frontend	3	10	\N	\N
f8f165a1-37b6-4a1b-b177-fa08b3149906	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	Tailwind CSS	Frontend	3	11	\N	\N
cf27b4c1-73c7-4e97-8b48-93b9675b3a3d	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	PHP	Backend & API	5	12	\N	\N
\.


--
-- Data for Name: social_links; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.social_links (id, person_id, platform_id, url, sort_order, created_at, updated_at) FROM stdin;
deb189fc-dcf2-41c2-a1f9-476a7464c65a	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	b2a2b0e0-e1e5-4b8c-b9b9-6b3a3c1e2b3c	https://github.com/lonelyman	1	\N	\N
a87d4108-091d-4833-8bbe-db1e6e2f78ad	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	c3c9310a-86cc-4495-9c60-080db387e3b1	https://www.facebook.com/niponxyz	2	2026-04-21 09:14:54.856701+00	2026-04-21 09:14:54.856701+00
\.


--
-- Data for Name: social_platforms; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.social_platforms (id, key, name, icon_url, sort_order, is_active, created_at, updated_at) FROM stdin;
b2a2b0e0-e1e5-4b8c-b9b9-6b3a3c1e2b3c	github	GitHub	github-icon	0	t	\N	\N
c3c9310a-86cc-4495-9c60-080db387e3b1	facebook	Facebook	\N	2	t	2026-04-21 09:14:49.780391+00	2026-04-21 09:14:49.780391+00
\.


--
-- Data for Name: user_auths; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_auths (id, user_id, provider, provider_id, secret, refresh_token) FROM stdin;
8e29ca88-21ce-4d09-8892-d472c41dce37	40b43774-a33c-4fd5-aa25-4eb98b5549a3	local	\N	$2a$10$ST12uB96VHRPvqqXFoKlEua0hpfAg677lq60wFYILsgqoznokijKy	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, person_id, username, email, status, last_login_at, is_email_verified, email_verification_token, email_verification_token_expires_at, password_reset_token, password_reset_token_expires_at, created_at, updated_at, deleted_at, role) FROM stdin;
40b43774-a33c-4fd5-aa25-4eb98b5549a3	019d71bd-1c2d-7a2c-a1d7-fd84416611f9	nipon	niponxyz@gmail.com	active	2026-04-20 11:29:08.060311+00	t	\N	\N	\N	\N	2026-04-20 03:46:06.087458+00	2026-04-20 11:29:08.060669+00	\N	user
\.


--
-- Name: experiences experiences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.experiences
    ADD CONSTRAINT experiences_pkey PRIMARY KEY (id);


--
-- Name: persons persons_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.persons
    ADD CONSTRAINT persons_pkey PRIMARY KEY (id);


--
-- Name: portfolio_item_tags portfolio_item_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.portfolio_item_tags
    ADD CONSTRAINT portfolio_item_tags_pkey PRIMARY KEY (portfolio_item_id, tag_id);


--
-- Name: portfolio_items portfolio_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.portfolio_items
    ADD CONSTRAINT portfolio_items_pkey PRIMARY KEY (id);


--
-- Name: portfolio_tags portfolio_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.portfolio_tags
    ADD CONSTRAINT portfolio_tags_pkey PRIMARY KEY (id);


--
-- Name: profiles profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.profiles
    ADD CONSTRAINT profiles_pkey PRIMARY KEY (id);


--
-- Name: skills skills_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skills
    ADD CONSTRAINT skills_pkey PRIMARY KEY (id);


--
-- Name: social_links social_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.social_links
    ADD CONSTRAINT social_links_pkey PRIMARY KEY (id);


--
-- Name: social_platforms social_platforms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.social_platforms
    ADD CONSTRAINT social_platforms_pkey PRIMARY KEY (id);


--
-- Name: user_auths user_auths_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_auths
    ADD CONSTRAINT user_auths_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_experiences_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_experiences_person_id ON public.experiences USING btree (person_id);


--
-- Name: idx_persons_created_by_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_persons_created_by_user_id ON public.persons USING btree (created_by_user_id);


--
-- Name: idx_persons_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_persons_deleted_at ON public.persons USING btree (deleted_at);


--
-- Name: idx_persons_invite_email_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_persons_invite_email_active ON public.persons USING btree (invite_email) WHERE (deleted_at IS NULL);


--
-- Name: idx_persons_invite_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_persons_invite_token ON public.persons USING btree (invite_token) WHERE (deleted_at IS NULL);


--
-- Name: idx_persons_linked_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_persons_linked_user ON public.persons USING btree (linked_user_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_portfolio_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_portfolio_items_deleted_at ON public.portfolio_items USING btree (deleted_at);


--
-- Name: idx_portfolio_items_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_portfolio_items_person_id ON public.portfolio_items USING btree (person_id);


--
-- Name: idx_portfolio_tags_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_portfolio_tags_name ON public.portfolio_tags USING btree (name);


--
-- Name: idx_portfolio_tags_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_portfolio_tags_slug ON public.portfolio_tags USING btree (slug);


--
-- Name: idx_profiles_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_profiles_person_id ON public.profiles USING btree (person_id);


--
-- Name: idx_provider_provider_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_provider_provider_id ON public.user_auths USING btree (provider_id);


--
-- Name: idx_skills_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_skills_person_id ON public.skills USING btree (person_id);


--
-- Name: idx_social_links_person_platform; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_social_links_person_platform ON public.social_links USING btree (person_id, platform_id);


--
-- Name: idx_social_platforms_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_social_platforms_key ON public.social_platforms USING btree (key);


--
-- Name: idx_user_auths_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_auths_user_id ON public.user_auths USING btree (user_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email_active ON public.users USING btree (email) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_email_verification_token_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email_verification_token_active ON public.users USING btree (email_verification_token) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_password_reset_token_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_password_reset_token_active ON public.users USING btree (password_reset_token) WHERE (deleted_at IS NULL);


--
-- Name: idx_users_person; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_person ON public.users USING btree (person_id) WHERE (deleted_at IS NULL);


--
-- Name: portfolio_item_tags fk_portfolio_item_tags_portfolio_item_model; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.portfolio_item_tags
    ADD CONSTRAINT fk_portfolio_item_tags_portfolio_item_model FOREIGN KEY (portfolio_item_id) REFERENCES public.portfolio_items(id);


--
-- Name: portfolio_item_tags fk_portfolio_item_tags_portfolio_tag_model; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.portfolio_item_tags
    ADD CONSTRAINT fk_portfolio_item_tags_portfolio_tag_model FOREIGN KEY (tag_id) REFERENCES public.portfolio_tags(id);


--
-- Name: social_links fk_profiles_social_links; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.social_links
    ADD CONSTRAINT fk_profiles_social_links FOREIGN KEY (person_id) REFERENCES public.profiles(person_id);


--
-- Name: user_auths fk_users_auths; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_auths
    ADD CONSTRAINT fk_users_auths FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

\unrestrict hFDahcxSxpDfZoCzdr4dRZeCD1PY303I2XnG5QeOUbwHxsX6SJG3seYFAkyZwhN

