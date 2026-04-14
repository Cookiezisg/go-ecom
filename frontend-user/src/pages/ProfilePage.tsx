import { FormEvent, useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getUserInfo, getUserCoupons, updateUserInfo, uploadFile } from "@/api/store";
import { useAuthStore } from "@/stores/auth";

export function ProfilePage() {
  const { profile, setAuth, token } = useAuthStore();
  const [message, setMessage] = useState("");
  const infoQuery = useQuery({
    queryKey: ["profile-info", profile?.id],
    queryFn: () => getUserInfo(profile!.id),
    enabled: Boolean(profile?.id),
  });

  const couponsQuery = useQuery({
    queryKey: ["profile-coupons", profile?.id],
    queryFn: () => getUserCoupons(profile!.id, 0),
    enabled: Boolean(profile?.id),
  });

  const updateMutation = useMutation({
    mutationFn: updateUserInfo,
    onSuccess: (response) => {
      if (profile) {
        setAuth(token, { ...profile, ...response.data });
      }
      setMessage("资料已更新");
    },
  });
  const uploadMutation = useMutation({
    mutationFn: (file: File) => uploadFile(file, "image"),
    onSuccess: (response) => {
      const url = response.data?.file_url;
      if (url && currentProfile) {
        setAuth(token, { ...currentProfile, avatar: url });
      }
    },
  });

  useEffect(() => {
    if (infoQuery.data?.data && profile) {
      setAuth(token, { ...profile, ...infoQuery.data.data });
    }
  }, [infoQuery.data, profile, setAuth, token]);

  if (!profile) {
    return null;
  }

  const currentProfile = profile;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    updateMutation.mutate({
      user_id: currentProfile.id,
      nickname: String(formData.get("nickname") || ""),
      avatar: String(formData.get("avatar") || ""),
      gender: Number(formData.get("gender") || 0),
      birthday: String(formData.get("birthday") || ""),
    });
  }

  return (
    <section className="stack">
      <div className="section-head">
        <h1>个人中心</h1>
        <div className="hero-actions">
          <Link className="secondary-button" to="/addresses">
            地址管理
          </Link>
          <Link className="secondary-button" to="/coupons">
            我的优惠券
          </Link>
        </div>
      </div>
      <div className="stack two-columns">
        <div className="panel">
          <h2>基本资料</h2>
        <form className="form" onSubmit={handleSubmit}>
          <label>
            用户名
            <input defaultValue={currentProfile.username} disabled />
          </label>
          <label>
            昵称
            <input defaultValue={currentProfile.nickname || ""} name="nickname" />
          </label>
            <label>
              头像 URL
              <input defaultValue={currentProfile.avatar || ""} name="avatar" />
            </label>
            <label>
              上传头像
              <input
                accept="image/*"
                onChange={(event) => {
                  const file = event.target.files?.[0];
                  if (file) {
                    uploadMutation.mutate(file);
                  }
                }}
                type="file"
              />
            </label>
          <label>
            性别
            <select defaultValue={String(currentProfile.gender || 0)} name="gender">
                <option value="0">未知</option>
                <option value="1">男</option>
                <option value="2">女</option>
              </select>
            </label>
          <label>
            生日
            <input defaultValue={currentProfile.birthday || ""} name="birthday" placeholder="YYYY-MM-DD" />
            </label>
            {message ? <div className="success-box">{message}</div> : null}
            <button className="primary-button" type="submit">
              保存资料
            </button>
          </form>
        </div>
        <div className="panel">
          <h2>账户概览</h2>
          <div className="stack compact">
            <div className="cart-row">
              <span>会员等级</span>
              <strong>{currentProfile.member_level ?? 0}</strong>
            </div>
            <div className="cart-row">
              <span>积分</span>
              <strong>{currentProfile.points ?? 0}</strong>
            </div>
            <div className="cart-row">
              <span>可用优惠券</span>
              <strong>{couponsQuery.data?.data?.length ?? 0}</strong>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
