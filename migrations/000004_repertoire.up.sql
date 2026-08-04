CREATE TABLE pieces (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    composer TEXT NOT NULL,
    category TEXT NOT NULL,
    difficulty INT NOT NULL CHECK(difficulty BETWEEN 1 AND 10),
    phase INT NOT NULL DEFAULT 1,
    xp_reward INT NOT NULL DEFAULT 200,
    imslp_url TEXT,
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE piece_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    piece_id INT NOT NULL REFERENCES pieces(id),
    status TEXT NOT NULL DEFAULT 'not_started'
        CHECK(status IN ('not_started','learning','mastering','conquered')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, piece_id)
);
