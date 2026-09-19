import "./user-badge.css";
import "./sport.css";
import "./sport-tabs.css";
import { lazy, Suspense, useEffect, useRef, useState } from "react";
import { Link, NavLink, Navigate, Route, Routes, useLocation } from "react-router-dom";
import { getCurrentUser } from "./api";
import { FoodFormPage } from "./pages/FoodFormPage";
import { SportFormPage } from "./pages/SportPage";
import { SportCatalogPage } from "./pages/SportCatalogPage";
import { SportActivityOverviewPage } from "./pages/SportActivityOverviewPage";
import { SportActivityFormPage } from "./pages/SportActivityFormPage";
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
  { path: "/weight", label: "Вес" },
  { path: "/sport", label: "Спорт" },
  { path: "/sport-activity", label: "Активность" }
];

export function App() {
  const [userName, setUserName] = useState("Пользователь");
  const [navigationOpen, setNavigationOpen] = useState(false);
  const navigationRef = useRef<HTMLDivElement>(null);
  const location = useLocation();
  const currentSection = location.pathname === "/settings"
    ? "Настройки"
    : sections.find((section) => location.pathname === section.path || location.pathname.startsWith(section.path + "/"))?.label ?? "Журнал";

  useEffect(() => {
    void getCurrentUser().then((user) => user?.name && setUserName(user.name)).catch(() => undefined);
  }, []);

  useEffect(() => {
    setNavigationOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!navigationOpen) return;
    function closeOnOutsideClick(event: PointerEvent) {
      if (event.target instanceof Node && !navigationRef.current?.contains(event.target)) setNavigationOpen(false);
    }
    function closeOnEscape(event: KeyboardEvent) {
      if (event.key === "Escape") setNavigationOpen(false);
    }
    document.addEventListener("pointerdown", closeOnOutsideClick);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnOutsideClick);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [navigationOpen]);

  return (
    <div className="app-shell">
      <header className="topbar">
        <NavLink className="brand" to="/">
          <span className="brand-mark">M+</span><span>MyHealth</span>
        </NavLink>
        <div className={`main-navigation ${navigationOpen ? "is-open" : ""}`} ref={navigationRef}>
          <button
            aria-controls="primary-navigation"
            aria-expanded={navigationOpen}
            aria-label="Выбрать раздел"
            className="mobile-nav-toggle"
            onClick={() => setNavigationOpen((open) => !open)}
            type="button"
          >
            <span>{currentSection}</span>
            <span aria-hidden="true" className="mobile-nav-chevron">⌄</span>
          </button>
          <nav aria-label="Основная навигация" id="primary-navigation">
            {sections.map((section) => (
              <NavLink key={section.path} onClick={() => setNavigationOpen(false)} to={section.path}>{section.label}</NavLink>
            ))}
          </nav>
        </div>
        <Link className="user-badge" to="/settings" aria-label={`Настройки пользователя: ${userName}`} title="Открыть настройки пользователя">
          <span aria-hidden="true">{userName.slice(0, 1).toUpperCase()}</span>
          <strong>{userName}</strong>
        </Link>
      </header>
      <main>
        <Routes>
          <Route path="/food" element={<FoodPage />} />
          <Route path="/" element={<Navigate replace to="/journal" />} />
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
          <Route path="/sport" element={<SportCatalogPage />} />
          <Route path="/sport/new" element={<SportFormPage />} />
          <Route path="/sport/:key/edit" element={<SportFormPage />} />
          <Route path="/sport-activity" element={<SportActivityOverviewPage />} />
          <Route path="/sport-activity/new" element={<SportActivityFormPage />} />
          <Route path="/sport-activity/:dt/:sportKey/edit" element={<SportActivityFormPage />} />
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
