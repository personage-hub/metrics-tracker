CREATE TABLE IF NOT EXISTS gauges (
                                      id SERIAL PRIMARY KEY,
                                      name VARCHAR(255) NOT NULL UNIQUE,
                                      value DOUBLE PRECISION NOT NULL,
                                      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS counters (
                                        id SERIAL PRIMARY KEY,
                                        name VARCHAR(255) NOT NULL UNIQUE,
                                        value BIGINT NOT NULL,
                                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                        updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
