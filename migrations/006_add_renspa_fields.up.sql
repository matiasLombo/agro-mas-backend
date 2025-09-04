-- Agregar campos RENSPA al modelo User
ALTER TABLE users ADD COLUMN renspa VARCHAR(20);
ALTER TABLE users ADD COLUMN establishment_name VARCHAR(200);
ALTER TABLE users ADD COLUMN establishment_location VARCHAR(300);

-- Agregar índice en RENSPA para búsquedas rápidas
CREATE INDEX IF NOT EXISTS idx_users_renspa ON users(renspa);

-- Comentarios para documentar los campos
COMMENT ON COLUMN users.renspa IS 'Código RENSPA - Registro Nacional Sanitario de Productores Agropecuarios';
COMMENT ON COLUMN users.establishment_name IS 'Nombre del establecimiento agropecuario';
COMMENT ON COLUMN users.establishment_location IS 'Ubicación del establecimiento agropecuario';