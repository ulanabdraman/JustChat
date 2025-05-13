CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       type VARCHAR(10) CHECK (type IN ('admin', 'user', 'bot')),
                       username TEXT UNIQUE NOT NULL,
                       online BOOLEAN DEFAULT false,
                       was_online TIMESTAMP,
                       created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE chats (
                       id SERIAL PRIMARY KEY,
                       name TEXT,
                       type VARCHAR(10) CHECK (type IN ('dialog', 'group', 'bot')),
                       created_by INTEGER REFERENCES users(id),
                       created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE chat_members (
                              chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
                              user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                              PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages (
                          id SERIAL PRIMARY KEY,
                          chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
                          creator_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
                          text TEXT,
                          sent_at TIMESTAMP DEFAULT now(),
                          deleted BOOLEAN DEFAULT false
);
