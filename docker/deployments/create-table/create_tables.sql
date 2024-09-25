CREATE SCHEMA account;

create schema schedule;

create schema teacher_info;

create schema adress_contacts;

CREATE TABLE account.type_role (
    role_name text primary key
);

CREATE TABLE account.profile (
    profile_tg_id bigint PRIMARY KEY, --on delete cascade on update cascade,
    profile_role_name TEXT NOT NULL, -- Новое поле для роли пользователя
    profile_name varchar(100) not null,

    foreign key (profile_role_name) references account.type_role(role_name)
);

--номер группы
create table schedule.type_group (
    type_group_name text primary key 
);

--группа (многие ко многим)
create table schedule.group (
    group_type_group_name text,
    group_profile_tg_id bigint,
    foreign key (group_type_group_name) references schedule.type_group(type_group_name),
    foreign key (group_profile_tg_id) references account.profile(profile_tg_id),
    primary key (group_type_group_name, group_profile_tg_id)
);

--вид занятий (лекция\лаба)
create table schedule.type_education (
    type_education_name text primary key 
);

--место проведения занятий ("онлайн", если ссылка)
create table schedule.type_education_room (
    type_education_room_name text primary key
);

create table schedule.subject (
    subject_name text primary key
);

CREATE TABLE teacher_info.type_department ( --кафедра
    type_department_name varchar(250) primary key
);

CREATE TABLE teacher_info.teacher (
    teacher_profile_tg_id bigint PRIMARY KEY, -- tg id
    teacher_name varchar(250) NOT NULL, -- фио
    teacher_job varchar(250) NOT NULL, -- должность
    teacher_department varchar(250) NOT NULL, -- кафедра
    teacher_adress varchar(250), -- адрес
    teacher_email varchar(250) UNIQUE, -- email
    FOREIGN KEY (teacher_profile_tg_id) REFERENCES account.profile(profile_tg_id),
    FOREIGN KEY (teacher_department) REFERENCES teacher_info.type_department(type_department_name)
);

CREATE TABLE teacher_info.image (
    teacher_id bigint PRIMARY KEY,
    name_img text,
    FOREIGN KEY (teacher_id) REFERENCES teacher_info.teacher(teacher_profile_tg_id)
);

--занятие
create table schedule.lesson (                    
    lesson_id serial primary key,                                              
    
    lesson_type_group_name text not null,           --название группы (лучше так при нанышней реализации)
    lesson_type_education_name text not null,                                    
    lesson_type_education_room_name text not null,   
    lesson_type_education_room_link text,           --ссылка на лекцию, если очно, то пусто       
    lesson_date_time_start timestamp not null,      --время_дата проведения ссылки
    lesson_date_time_end timestamp not null,                  
    lesson_link text,                               --ссылка, если онлик
    lesson_teacher_id bigint,                                                     
    lesson_subject_name text not null,

    foreign key (lesson_teacher_id) references account.profile(profile_tg_id), --id препода

    foreign key (lesson_subject_name) references schedule.subject(subject_name),
    foreign key (lesson_type_group_name) references schedule.type_group(type_group_name),    
    foreign key (lesson_type_education_name) references schedule.type_education(type_education_name),                                                                                      
    foreign key (lesson_type_education_room_name) references schedule.type_education_room(type_education_room_name)                                                                          
);

CREATE TABLE adress_contacts.place (
    place_name varchar(250) primary key,
    place_time_start time not null,
    place_time_end time not null,
    place_phone varchar(12),
    place_email varchar(250),
    place_adress varchar(250) not null,
    place_latitude double precision not null,
    place_longitude double precision not null
);

CREATE TABLE adress_contacts.type_place (
    type_place_name varchar(250) primary key
);

CREATE TABLE adress_contacts.place_info (
    place_info_id serial primary key,
    place_info_type_place_name varchar(250),
    place_info_place_name varchar(250),

    foreign key (place_info_type_place_name) references adress_contacts.type_place(type_place_name),                                                                                      
    foreign key (place_info_place_name) references adress_contacts.place(place_name)  
);
