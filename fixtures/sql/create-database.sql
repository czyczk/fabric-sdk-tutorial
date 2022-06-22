CREATE DATABASE fabric_phone CHARACTER SET utf8mb4;

GRANT ALL PRIVILEGES ON fabric_phone.* TO 'fabricphone' @'%';

CREATE DATABASE fabric_phone_ado1 CHARACTER SET utf8mb4;

GRANT ALL PRIVILEGES ON fabric_phone_ado1.* TO 'fabricphone' @'%';

CREATE DATABASE fabric_phone_u1o2 CHARACTER SET utf8mb4;

GRANT ALL PRIVILEGES ON fabric_phone_u1o2.* TO 'fabricphone' @'%';

FLUSH PRIVILEGES;