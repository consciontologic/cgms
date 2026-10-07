ALTER TABLE rooms ADD COLUMN bot_difficulty text NOT NULL DEFAULT ''
 CHECK(bot_difficulty IN ('','beginner','standard','advanced'));
