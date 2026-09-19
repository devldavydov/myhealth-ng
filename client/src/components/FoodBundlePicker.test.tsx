import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { FoodBundlePicker } from "./FoodBundlePicker";

afterEach(()=>{cleanup();vi.unstubAllGlobals();});
function response(body:unknown){return{ok:true,json:async()=>body};}

it("рендерит выпадающее меню вне обрезающего контейнера журнала",async()=>{
 const food={key:"food",name:"Творог",brand:"",cal100:100,prot100:10,fat100:5,carb100:3,comment:""};
 vi.stubGlobal("fetch",vi.fn().mockImplementation(async(input:string)=>input.startsWith("/api/food?")?response({data:[food],pagination:{page:1,pageSize:50,total:1,totalPages:1}}):response({data:[],pagination:{page:1,pageSize:50,total:0,totalPages:0}})));
 render(<div data-testid="clipping-zone" style={{overflow:"hidden"}}><FoodBundlePicker onError={()=>undefined} onSelect={()=>undefined}/></div>);
 const picker=screen.getByRole("combobox",{name:"Поиск еды или бандла"});fireEvent.focus(picker);fireEvent.change(picker,{target:{value:"Твор"}});await screen.findByText("Творог");
 const portal=document.querySelector<HTMLElement>(".bundle-select__menu-portal");expect(portal).not.toBeNull();expect(portal?.parentElement).toBe(document.body);expect(portal).toHaveStyle({position:"absolute"});expect(screen.getByTestId("clipping-zone")).not.toContainElement(portal);
});
