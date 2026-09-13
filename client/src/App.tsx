import "./user-badge.css";
import { useEffect, useState } from "react";
import { NavLink, Navigate, Route, Routes } from "react-router-dom";
import { getCurrentUser } from "./api";
import { DashboardPage } from "./pages/DashboardPage";
import { MeasurementsPage } from "./pages/MeasurementsPage";

export function App() {
  const [userName, setUserName] = useState("Пользователь");

  useEffect(() => {
    void getCurrentUser().then((user) => user?.name && setUserName(user.name)).catch(() => undefined);
  }, []);

  return (
    <div className="app-shell">
      <header className="topbar">
        <NavLink className="brand" to="/">
          <span className="brand-mark">M+</span><span>MyHealth</span>
        </NavLink>
        <nav aria-label="Основная навигация">
          <NavLink to="/">Обзор</NavLink>
          <NavLink to="/measurements">Измерения</NavLink>
        </nav>
        <div className="user-badge" title="Имя из клиентского сертификата">
          <span aria-hidden="true">{userName.slice(0, 1).toUpperCase()}</span>
          <strong>{userName}</strong>
        </div>
      </header>
      <main>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/measurements" element={<MeasurementsPage />} />
          <Route path="*" element={<Navigate replace to="/" />} />
        </Routes>
      </main>
    </div>
  );
}
