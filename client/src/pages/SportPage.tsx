import { type FormEvent, useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { createSport, getSport, getSports, updateSport, type Sport, type SportData } from "../api";

export function SportPage() {
  const [items, setItems] = useState<Sport[]>([]);
  const [error, setError] = useState("");
  useEffect(() => { void getSports().then((page) => setItems(page.items)).catch((e: unknown) => setError(e instanceof Error ? e.message : "Не удалось загрузить спорт")); }, []);
  return <div className="page"><header className="page-header"><div><p className="eyebrow">Справочник</p><h1>Спорт</h1><p className="page-description">Виды активности и единицы измерения.</p></div><Link className="button primary" to="/sport/new">Добавить вид</Link></header>{error && <div className="error">{error}</div>}<div className="sport-cards">{items.map((sport) => <article className="food-card" key={sport.key}><div className="food-card-heading"><h3>{sport.name}</h3><strong><small>{sport.unit}</small></strong></div>{sport.comment && <p className="food-comment">{sport.comment}</p>}<div className="food-actions"><Link className="text-button" to={`/sport/${encodeURIComponent(sport.key)}/edit`}>Редактировать</Link></div></article>)}</div></div>;
}
export function SportFormPage() {
  const { key } = useParams(); const navigate = useNavigate(); const editing = Boolean(key);
  const [data, setData] = useState<SportData>({ name: "", unit: "", comment: "" }); const [error,setError]=useState("");
  useEffect(()=>{ if(key) void getSport(key).then(setData).catch((e:unknown)=>setError(e instanceof Error?e.message:"Не удалось загрузить вид спорта")); },[key]);
  async function submit(event: FormEvent) { event.preventDefault(); try { if(key) await updateSport(key,data); else await createSport(data); navigate("/sport"); } catch(e) { setError(e instanceof Error?e.message:"Не удалось сохранить вид спорта"); } }
  return <div className="page food-form-page"><Link className="back-link" to="/sport">← К спорту</Link><header className="form-page-header"><p className="eyebrow">{editing?"Редактирование":"Новый вид"}</p><h1>{editing?"Изменить вид спорта":"Добавить вид спорта"}</h1></header><section className="food-form-card">{error&&<div className="error">{error}</div>}<form onSubmit={submit}><label>Название<input value={data.name} onChange={e=>setData({...data,name:e.target.value})}/></label><label>Единица измерения<input value={data.unit} placeholder="Например, шт или км" onChange={e=>setData({...data,unit:e.target.value})}/></label><label>Комментарий<textarea value={data.comment} onChange={e=>setData({...data,comment:e.target.value})}/></label><div className="form-actions"><button className="button primary">Сохранить</button><Link className="button secondary" to="/sport">Отмена</Link></div></form></section></div>;
}
