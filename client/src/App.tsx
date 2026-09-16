import "./user-badge.css";
import { lazy, Suspense, useEffect, useState } from "react";
import { Link, NavLink, Navigate, Route, Routes } from "react-router-dom";
import { getCurrentUser } from "./api";
import { FoodFormPage } from "./pages/FoodFormPage";
import { FoodPage } from "./pages/FoodPage";

const WeightPage = lazy(() => import("./pages/WeightPage").then((module) => ({ default: module.WeightPage })));
const BundlePage = lazy(() => import("./pages/BundlePage").then((module) => ({ default: module.BundlePage })));
const BundleFormPage = lazy(() => import("./pages/BundleFormPage").then((module) => ({ default: module.BundleFormPage })));
const SettingsPage = lazy(() => import("./pages/SettingsPage").then((module) => ({ default: module.SettingsPage })));
const JournalPage = lazy(() => import("./pages/JournalPage").then((module) => ({ default: module.JournalPage })));

const sections = [
  { path: "/journal", label: "Журнал" },
  { path: "/food", label: "Еда" },
  { path: "/bundle", label: "Бандлы" },
  { path: "/weight", label: "Вес" }
];

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
          {sections.map((section) => (
            <NavLink key={section.path} to={section.path}>{section.label}</NavLink>
          ))}
        </nav>
        <Link className="user-badge" to="/settings" aria-label={`Настройки пользователя: ${userName}`} title="Открыть настройки пользователя">
          <span aria-hidden="true">{userName.slice(0, 1).toUpperCase()}</span>
          <strong>{userName}</strong>
        </Link>
      </header>
      <main>
        <Routes>
          <Route path="/food" element={<FoodPage />} />
          <Route path="/journal" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка журнала…</p></div>}>
              <JournalPage />
            </Suspense>
          )} />
          <Route path="/food/new" element={<FoodFormPage />} />
          <Route path="/food/:key/edit" element={<FoodFormPage />} />
          <Route path="/bundle" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка раздела…</p></div>}>
              <BundlePage />
            </Suspense>
          )} />
          <Route path="/bundle/new" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка формы…</p></div>}>
              <BundleFormPage />
            </Suspense>
          )} />
          <Route path="/bundle/:key/edit" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка формы…</p></div>}>
              <BundleFormPage />
            </Suspense>
          )} />
          <Route path="/weight" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка раздела…</p></div>}>
              <WeightPage />
            </Suspense>
          )} />
          <Route path="/settings" element={(
            <Suspense fallback={<div className="page"><p className="muted">Загрузка настроек…</p></div>}>
              <SettingsPage userName={userName} />
            </Suspense>
          )} />
          <Route path="*" element={<Navigate replace to="/food" />} />
        </Routes>
      </main>
    </div>
  );
}
