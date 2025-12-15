import { Routes, Route, Navigate } from "react-router-dom";
import Header from "./components/Header";
import TopicsPage from "./pages/TopicsPage";
import PostsPage from "./pages/PostsPage";
import PostPage from "./pages/PostPage";

export default function App() {
    return (
        <div>
            <Header />
            <div style={{ padding: 16, maxWidth: 900, margin: "0 auto" }}>
                <Routes>
                    <Route path="/" element={<Navigate to="/topics" replace />} />
                    <Route path="/topics" element={<TopicsPage />} />
                    <Route path="/topics/:topicID/posts" element={<PostsPage />} />
                    <Route path="/posts/:postID" element={<PostPage />} />
                    <Route path="*" element={<div>404</div>} />
                </Routes>
            </div>
        </div>
    );
}
