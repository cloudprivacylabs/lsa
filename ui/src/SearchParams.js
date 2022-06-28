import React from "react";
import { useState, useEffect } from "react";

const SearchParams = () => {
    const [schema, setSchema] = useState("");
    const [valueType, setValueType] = useState("");
    const [labelText, setLabelText] = useState("");

    useEffect(() => {}, []);

    return (
        <div className="search-params">
            <form
                onSubmit={(e) => {
                    e.preventDefault();
                }}
            >
                <label htmlFor="schema">
                    Schema
                    <input
                        id="schema"
                        placeholder="Schema"
                        value={schema}
                        onChange={(e) => setSchema(e.target.value)}
                    />
                </label>
                <label htmlFor="valueType">
                    valueType
                    <input
                        id="valueType"
                        value={valueType}
                        onChange={(e) => setValueType(e.target.value)}
                    ></input>
                </label>
                <label htmlFor="labelText">
                    valueType
                    <input
                        id="labelText"
                        value={labelText}
                        onChange={(e) => setLabelText(e.target.value)}
                    ></input>
                </label>
                <button>Generate</button>
            </form>
        </div>
    );
};

export default SearchParams;
