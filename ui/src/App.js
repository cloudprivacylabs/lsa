import React from "react";
import ReactDOM from "react-dom/client";
import { StrictMode } from "react";
import { BrowserRouter, Route, Routes, Link } from "react-router-dom";
import SearchParams from "./SearchParams";

const App = () => {
    return (
        <StrictMode>
            <BrowserRouter>
                <header>
                    {/* links must always be inside of the router */}
                    <Link to="/">Layered Schemas</Link>
                </header>
                <Routes>
                    <Route path="/" element={<SearchParams />} />
                </Routes>
            </BrowserRouter>
        </StrictMode>
    );
};

const root = ReactDOM.createRoot(document.getElementById("root"));
root.render(<App />);
