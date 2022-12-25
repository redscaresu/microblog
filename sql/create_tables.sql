-- Creation of blog table
CREATE TABLE IF NOT EXISTS blog (
  blog_id INT NOT NULL,
  blog_post varchar(10000) NOT NULL,
  PRIMARY KEY (blog_id)
);