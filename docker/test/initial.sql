INSERT INTO account.profile (profile_email, profile_phone, profile_hash_pass, profile_nickname, profile_role)
VALUES
    ('john.doe@example.com', 1234567890, '7f57b9267973b690a1133c3067000433017ca5f45d36e05f3a17602099a75098', 'johndoe', 'admin'),
    --hash_password_1
    ('jane.smith@example.com', 0987654321, 'e4f721d6adecf940b2711f630b76c55b49ef220cc4393d5c191faa613bd42893', 'janesmith', 'student'),
    --hash_password_2
    ('bob.johnson@example.com', 5555555555, '7fc6cb50cceb86a9f9fb62ef636f5fc0209ea6a0e2ca19bd3e408271e07b537b', 'bobjohnson', 'employee');
    --hash_password_3
   ;

select * from account.profile 
;
