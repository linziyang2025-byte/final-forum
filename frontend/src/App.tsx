import { Routes, Route, Navigate } from "react-router-dom";
import Header from "./components/Header";
import TopicsPage from "./pages/TopicsPage";
import PostsPage from "./pages/PostsPage";
import PostPage from "./pages/PostPage";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import { RequireAuth } from "./auth";

export default function App() {
    return (
        <div>
            <Header />
            <div className="page">
                <Routes>
                    <Route path="/" element={<Navigate to="/topics" replace />} />
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/register" element={<RegisterPage />} />
                    <Route path="/topics" element={<RequireAuth><TopicsPage /></RequireAuth>} />
                    <Route path="/topics/:topicID/posts" element={<RequireAuth><PostsPage /></RequireAuth>} />
                    <Route path="/posts/:postID" element={<RequireAuth><PostPage /></RequireAuth>} />
                    <Route path="*" element={<div>404</div>} />
                </Routes>
            </div>
        </div>
    );
}
