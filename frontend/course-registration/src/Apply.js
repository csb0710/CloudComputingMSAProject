import React, { useEffect, useState } from 'react';
import axios from 'axios';

const CourseTable = () => {
  const [courses, setCourses] = useState([]);
  const [enrolledCourses, setEnrolledCourses] = useState([]);

  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const response = await axios.get('http://localhost:8081/courses');
        if (!response.status === 200) {
          throw new Error('Failed to fetch courses');
        }
        setCourses(response.data);
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    fetchCourses();
  }, []);

  useEffect(() => {
    const storedEnrolledCourses = localStorage.getItem('enrolledCourses');
    if (storedEnrolledCourses) {
      setEnrolledCourses(JSON.parse(storedEnrolledCourses));
    }
  }, []);

  const handleRowClick = async (course) => {
    const enrollRequest = {
      student_id: "32194501",
      subj_id: course.subjId,
      dvcls_nb: course.dvclsNb,
    };

    try {
      const response = await fetch('http://localhost:8081/enroll', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(enrollRequest),
      });

      const data = await response.json();

      if (response.ok) {
        alert(`Course application successful: ${data.message}`);
        window.location.reload();
        const updatedEnrolledCourses = [...enrolledCourses, course];
        if (data.message !== "수강 신청이 마감된 강좌입니다") {
          setEnrolledCourses(updatedEnrolledCourses);
        }
        localStorage.setItem('enrolledCourses', JSON.stringify(updatedEnrolledCourses));
      } else {
        alert(`Course application failed: ${data.message}`);
      }
    } catch (error) {
      console.error('Error applying for course:', error);
      alert('An error occurred. Please try again.');
    }
  };

  const handleClearEnrolledCourses = () => {
    localStorage.removeItem('enrolledCourses');
    setEnrolledCourses([]);
  };

  return (
    <>
      <table>
        <thead>
          <tr>
            <th>교과목번호</th>
            <th>단과대학</th>
            <th>학년</th>
            <th>분류</th>
            <th>분반</th>
            <th>학점</th>
            <th>교과목명</th>
            <th>교강사명</th>
            <th>요일/교시/강의실</th>
            <th>신청인원</th>
            <th>제한인원</th>
          </tr>
        </thead>
        <tbody>
          {courses.map((course, index) => (
            <tr key={index} onClick={() => handleRowClick(course)}>
              <td>{course.subjId}</td>
              <td>{course.tkcrsEcaOrgnm}</td>
              <td>{course.grade}</td>
              <td>{course.curiCparNm}</td>   
              <td>{course.dvclsNb}</td>         
              <td>{course.crd}</td>
              <td>{course.subjKnm}</td>
              <td>{course.wkLecrEmpnm}</td>
              <td>{course.buldAndRoomCont}</td>
              <td>{course.enrolledStudents !== undefined ? course.enrolledStudents : 0}</td>
              <td>{course.maxStudents}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {enrolledCourses.length > 0 && (
        <div>
          <h2>Successfully Enrolled Courses</h2>
          <table>
            <thead>
              <tr>
                <th>교과목번호</th>
                <th>단과대학</th>
                <th>학년</th>
                <th>분류</th>
                <th>분반</th>
                <th>학점</th>
                <th>교과목명</th>
                <th>교강사명</th>
                <th>요일/교시/강의실</th>
              </tr>
            </thead>
            <tbody>
              {enrolledCourses.map((course, index) => (
                <tr key={index}>
                  <td>{course.subjId}</td>
                  <td>{course.tkcrsEcaOrgnm}</td>
                  <td>{course.grade}</td>
                  <td>{course.curiCparNm}</td>   
                  <td>{course.dvclsNb}</td>         
                  <td>{course.crd}</td>
                  <td>{course.subjKnm}</td>
                  <td>{course.wkLecrEmpnm}</td>
                  <td>{course.buldAndRoomCont}</td>
                </tr>
              ))}
            </tbody>
          </table>
          
        </div>
      )}

          <button 
            style={{ marginTop: '20px', padding: '10px', backgroundColor: '#f8f9fa', color: 'black', border: 'none', cursor: 'pointer' }}
            onClick={handleClearEnrolledCourses}
          >
          </button>
    </>
  );
};

export default CourseTable;
