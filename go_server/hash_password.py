#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
生成用户密码的SHA256哈希值
使用方式: python3 hash_password.py <password> <salt>
"""

import sys
import hashlib

def hash_password(password_md5, salt):
    """
    根据MD5密码和salt生成最终哈希值
    password_md5: 前端MD5哈希后的密码
    salt: 用户的盐值
    返回: SHA256(password_md5 + salt)
    """
    combined = password_md5 + salt
    hash_obj = hashlib.sha256(combined.encode('utf-8'))
    return hash_obj.hexdigest()

def md5_hash(password):
    """计算密码的MD5哈希"""
    return hashlib.md5(password.encode('utf-8')).hexdigest()

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("使用方式:")
        print("  python3 hash_password.py <password> <salt>")
        print("  python3 hash_password.py 123456 admin2026electric")
        print("")
        print("或直接传入MD5值:")
        print("  python3 hash_password.py --md5 <password_md5> <salt>")
        sys.exit(1)
    
    try:
        if sys.argv[1] == "--md5":
            # 直接使用MD5值
            if len(sys.argv) < 4:
                print("错误: --md5模式需要提供password_md5和salt两个参数")
                sys.exit(1)
            password_md5 = sys.argv[2]
            salt = sys.argv[3]
        else:
            # 明文密码，先计算MD5
            password = sys.argv[1]
            salt = sys.argv[2]
            password_md5 = md5_hash(password)
            print("密码: {}".format(password))
            print("MD5: {}".format(password_md5))
        
        print("Salt: {}".format(salt))
        
        # 计算最终哈希
        final_hash = hash_password(password_md5, salt)
        print("\n最终SHA256哈希: {}".format(final_hash))
        print("\nSQL语句:")
        print("UPDATE users_tab SET password='{}', salt='{}' WHERE email='admin@gmail.com';".format(final_hash, salt))
    except Exception as e:
        print("错误: {}".format(str(e)))
        sys.exit(1)

