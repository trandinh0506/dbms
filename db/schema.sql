CREATE TABLE ACCOUNT (
    id INT PRIMARY KEY IDENTITY(1,1),
    username NVARCHAR(50) UNIQUE,
    password_hash NVARCHAR(255),
    role NVARCHAR(20)
);

CREATE TABLE STUDENT (
    id INT PRIMARY KEY,
    name NVARCHAR(100),
    FOREIGN KEY (id) REFERENCES ACCOUNT(id)
);

CREATE TABLE INSTRUCTOR (
    id INT PRIMARY KEY,
    name NVARCHAR(100),
    FOREIGN KEY (id) REFERENCES ACCOUNT(id)
);

CREATE TABLE COURSE (
    id INT PRIMARY KEY IDENTITY(1,1),
    name NVARCHAR(200),
    capacity INT DEFAULT 30,
    created_by INT,
    FOREIGN KEY (created_by) REFERENCES INSTRUCTOR(id)
);

CREATE TABLE COURSE_INSTRUCTOR (
    course_id INT,
    instructor_id INT,
    PRIMARY KEY (course_id, instructor_id),
    FOREIGN KEY (course_id) REFERENCES COURSE(id),
    FOREIGN KEY (instructor_id) REFERENCES INSTRUCTOR(id)
);

CREATE TABLE ENROLLMENT (
    student_id INT,
    course_id INT,
    mark DECIMAL(4,2),
    updated_at DATETIME DEFAULT GETDATE(),
    PRIMARY KEY (student_id, course_id),
    FOREIGN KEY (student_id) REFERENCES STUDENT(id),
    FOREIGN KEY (course_id) REFERENCES COURSE(id)
);

-- 3. TẠO VIEW
CREATE OR ALTER VIEW View_CourseLecturers AS
SELECT 
    c.id AS CourseID,
    c.name AS CourseName, 
    ISNULL(STRING_AGG(i.name, ', '), N'Chưa có') AS Lecturers, 
    COUNT(DISTINCT e.student_id) AS StudentCount
FROM COURSE c
LEFT JOIN COURSE_INSTRUCTOR ci ON c.id = ci.course_id
LEFT JOIN INSTRUCTOR i ON ci.instructor_id = i.id
LEFT JOIN ENROLLMENT e ON c.id = e.course_id
GROUP BY c.id, c.name;


-- 4. TẠO FUNCTION
CREATE FUNCTION dbo.fn_GetStudentGPA (@studentId INT)
RETURNS DECIMAL(4,2) AS
BEGIN
    RETURN (SELECT ISNULL(AVG(mark), 0) FROM ENROLLMENT WHERE student_id = @studentId);
END;


-- 5. TẠO PROCEDURE
CREATE OR ALTER PROCEDURE RegisterAccount 
    @username NVARCHAR(50), 
    @password_hash NVARCHAR(255),
    @fullname NVARCHAR(100),
    @role NVARCHAR(20)
AS
BEGIN
    SET NOCOUNT ON;
    INSERT INTO ACCOUNT (username, password_hash, role) VALUES (@username, @password_hash, @role);
    DECLARE @new_id INT = SCOPE_IDENTITY();
    
    IF @role = 'Student'
        INSERT INTO STUDENT (id, name) VALUES (@new_id, @fullname);
    ELSE IF @role = 'Instructor'
        INSERT INTO INSTRUCTOR (id, name) VALUES (@new_id, @fullname);
        
    SELECT @new_id AS GeneratedID;
END;

-- dữ liệu rác
-- dirty read
CREATE OR ALTER PROCEDURE GetStudentGradesSafe
    @SID INT
ASread
BEGIN
    SET NOCOUNT ON;
    SET LOCK_TIMEOUT 100;

    DECLARE @Result TABLE (
        course_name NVARCHAR(255),
        mark DECIMAL(5,2)
    );

    DECLARE @CID INT, @CName NVARCHAR(255);
    
    DECLARE course_cursor CURSOR FOR 
    SELECT c.id, c.name 
    FROM ENROLLMENT e WITH (NOLOCK)
    JOIN COURSE c WITH (NOLOCK) ON e.course_id = c.id
    WHERE e.student_id = @SID;

    OPEN course_cursor;
    FETCH NEXT FROM course_cursor INTO @CID, @CName;

    WHILE @@FETCH_STATUS = 0
    BEGIN
        BEGIN TRY
            DECLARE @CurrentMark DECIMAL(5,2);
            
            SELECT @CurrentMark = mark FROM ENROLLMENT 
            WHERE student_id = @SID AND course_id = @CID;

            INSERT INTO @Result VALUES (@CName, @CurrentMark);
        END TRY
        BEGIN CATCH
            IF ERROR_NUMBER() = 1222
            BEGIN
                INSERT INTO @Result VALUES (@CName, -1);
            END
            ELSE
            BEGIN
                INSERT INTO @Result VALUES (@CName, -2);
            END
        END CATCH

        FETCH NEXT FROM course_cursor INTO @CID, @CName;
    END

    CLOSE course_cursor;
    DEALLOCATE course_cursor;

    SELECT * FROM @Result;
END;


--------

-- non-repeatable read
-- không đọc lại được dữ liệu trong cùng 1 transaction nếu có 1 transaction khác đã update dữ liệu đó

CREATE OR ALTER PROCEDURE FinalizeAcademicReport
    @CourseID INT,
    @IsSafe BIT 
AS
BEGIN
    SET NOCOUNT ON;
    IF @IsSafe = 1 
        SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
    ELSE 
        SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED; 

    BEGIN TRANSACTION;
    BEGIN TRY
        SELECT student_id, mark 
        FROM ENROLLMENT 
        WHERE course_id = @CourseID;

        WAITFOR DELAY '00:00:20';

        SELECT student_id, mark 
        FROM ENROLLMENT 
        WHERE course_id = @CourseID;

        COMMIT TRANSACTION;
    END TRY
    BEGIN CATCH
        IF @@TRANCOUNT > 0 ROLLBACK;
        THROW;
    END CATCH
END;

----------


-- phantom read
-- xuất hiện hoặc mất đi một/nhiều dòng dữ liệu trong cùng 1 transaction nếu có 1 transaction khác đã insert hoặc delete dữ liệu đó
CREATE OR ALTER PROCEDURE StatWeakStudentsReport
    @CourseID INT,
    @IsSafe BIT
AS
BEGIN
    SET NOCOUNT ON;

    IF @IsSafe = 1 
        SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
    ELSE 
        SET TRANSACTION ISOLATION LEVEL READ UNCOMMITTED;

    BEGIN TRANSACTION;
    BEGIN TRY
        SELECT student_id, mark 
        FROM ENROLLMENT 
        WHERE course_id = @CourseID AND mark < 5;

        WAITFOR DELAY '00:00:20';

        SELECT student_id, mark 
        FROM ENROLLMENT 
        WHERE course_id = @CourseID AND mark < 5;

        COMMIT TRANSACTION;
    END TRY
    BEGIN CATCH
        IF @@TRANCOUNT > 0 ROLLBACK;
        THROW;
    END CATCH
END;

----

CREATE OR ALTER PROCEDURE EnrollStudent
    @StudentID INT,
    @CourseID INT
AS
BEGIN
    SET NOCOUNT ON;
    
    BEGIN TRANSACTION;

    BEGIN TRY
        IF EXISTS (SELECT 1 FROM ENROLLMENT WHERE student_id = @StudentID AND course_id = @CourseID)
        BEGIN
            ROLLBACK;
            RAISERROR(N'Sinh viên đã đăng ký khóa học này rồi.', 16, 1);
            RETURN;
        END

        DECLARE @CurrentCount INT;
        DECLARE @MaxCapacity INT;

        SELECT @MaxCapacity = capacity FROM COURSE WITH (UPDLOCK) WHERE id = @CourseID;
        SELECT @CurrentCount = COUNT(*) FROM ENROLLMENT WHERE course_id = @CourseID;

        IF @CurrentCount >= @MaxCapacity
        BEGIN
            ROLLBACK;
            RAISERROR(N'Lớp học đã đầy (%d/%d). Không thể đăng ký.', 16, 1, @CurrentCount, @MaxCapacity);
            RETURN;
        END

        INSERT INTO ENROLLMENT (student_id, course_id, mark, updated_at)
        VALUES (@StudentID, @CourseID, 0, GETDATE());

        COMMIT TRANSACTION;
        PRINT 'Đăng ký thành công.';
    END TRY
    BEGIN CATCH
        IF @@TRANCOUNT > 0 ROLLBACK;
        DECLARE @ErrMsg NVARCHAR(4000) = ERROR_MESSAGE();
        RAISERROR(@ErrMsg, 16, 1);
    END CATCH
END;

--------


-- 6. TẠO TRIGGER
CREATE TRIGGER TRG_Enrollment_Update ON ENROLLMENT AFTER UPDATE AS
BEGIN
    IF (UPDATE(updated_at)) RETURN;
    IF EXISTS (SELECT 1 FROM inserted WHERE mark < 0 OR mark > 10)
    BEGIN
        RAISERROR (N'Lỗi điểm!', 16, 1);
        ROLLBACK;
    END
    UPDATE ENROLLMENT SET updated_at = GETDATE() FROM ENROLLMENT e JOIN inserted i ON e.student_id = i.student_id;
END;
