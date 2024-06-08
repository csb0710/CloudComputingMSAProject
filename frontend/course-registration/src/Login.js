import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './Login.css';

const Login = () => {
  const [studentId, setStudentId] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();

  const handleLogin = async (e) => {
    e.preventDefault();
    const response = await fetch('http://localhost:59198/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ studentId, password }),
    });

    if (response.ok) {
      // Handle successful login
      alert('Login successful!');
      navigate("/apply");
    } else {
      // Handle failed login
      alert('Login failed. Please check your credentials.');
    }
  };

  return (
    <div className="main-container">
      <div className="login-container">
        <div className="login-box">
          <h1>COURSE REGISTRATION SYSTEM</h1>
          <h2>단국대학교 수강신청시스템 로그인</h2>
          <form onSubmit={handleLogin} className="login-form">
            <div className="input-group">
              <input
                type="text"
                placeholder="학번"
                value={studentId}
                onChange={(e) => setStudentId(e.target.value)}
                required
              />
              <input
                type="password"
                placeholder="비밀번호"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
              <button type="submit">Login</button>
            </div>
          </form>
          <div className="links">
            <a href="#">아이디 찾기</a> | <a href="#">비밀번호 찾기</a>
          </div>
        </div>
      </div>
      <div className="info-container">
        <div className="info-box">
          <h3>불법 수강신청에 따른 안내</h3>
          <p>
            불법 수강신청으로 인하여 타학생에게 피해를 주고 대학 전산망에 악영향을 주는 행위에 대하여
            학내규정에서는 강력히 대응 할 것이며 학칙에 의거 엄중 처벌됨을 알려드립니다. 또한 수강신청 Black List에
            등록된 수강신청자는 향후 차기학기 수강신청 권한이 제한됨을 알려드립니다.
          </p>
          <ul>
            <li>타인의 비밀번호를 도용하여 수강신청하는 행위</li>
            <li>매크로 등 불법 프로그램을 이용하여 수강신청하는 행위</li>
            <li>기타 불법행위</li>
          </ul>
          <h3>수강신청 안내</h3>
          <p>
            최초 비밀번호는 주민등록번호 앞 10자리이며 웹정보시스템에서 비밀번호 변경이 가능합니다.
          </p>
          <p>
            수강신청 프로그램은 <span>Chrome, Firefox, Safari</span> 이용 가능 (브라우저별 최신 버전으로 업그레이드 권장)
          </p>
          <p>웹 수강신청 프로그램 이용 안내를 읽어보시기 바랍니다.</p>
        </div>
      </div>
    </div>
  );
};

export default Login;
