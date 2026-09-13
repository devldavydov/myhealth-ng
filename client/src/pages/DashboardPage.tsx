import { Link } from "react-router-dom";

const cards = [
  { label: "Самочувствие", value: "Хорошее", detail: "Последняя отметка — сегодня" },
  { label: "Пульс", value: "68", detail: "уд/мин · в пределах нормы" },
  { label: "Вес", value: "72,4", detail: "кг · −0,3 за неделю" }
];

export function DashboardPage() {
  return (
    <div className="page">
      <section className="hero">
        <div>
          <p className="eyebrow">13 сентября</p>
          <h1>Добрый день!</h1>
          <p>Все важные показатели здоровья — в одном спокойном месте.</p>
        </div>
        <Link className="button" to="/measurements">Добавить измерение</Link>
      </section>
      <section aria-labelledby="today-heading">
        <h2 id="today-heading">Сегодня</h2>
        <div className="card-grid">
          {cards.map((card) => (
            <article className="metric-card" key={card.label}>
              <p>{card.label}</p><strong>{card.value}</strong><span>{card.detail}</span>
            </article>
          ))}
        </div>
      </section>
      <section className="info-panel">
        <div>
          <p className="eyebrow">Следующий шаг</p>
          <h2>Сформируйте привычку</h2>
          <p>Регулярные записи помогают видеть динамику, а не отдельные цифры.</p>
        </div>
        <span className="progress" aria-label="Заполнено 4 дня из 7">4 / 7 дней</span>
      </section>
    </div>
  );
}
