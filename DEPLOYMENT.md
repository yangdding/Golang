# 배포 가이드

## Git 저장소 설정 및 배포

### 1. Git 초기화

```bash
# Git 저장소 초기화
git init

# 파일 추가
git add .

# 첫 커밋
git commit -m "Initial commit: File sharing service"
```

### 2. GitHub 저장소 연결

```bash
# GitHub에서 새 저장소 생성 후 아래 명령 실행
git remote add origin https://github.com/your-username/your-repo-name.git

# main 브랜치로 푸시
git branch -M main
git push -u origin main
```

### 3. Render 배포

1. **Render 대시보드 접속**
   - https://dashboard.render.com 접속
   - "New +" 버튼 클릭
   - "Web Service" 선택

2. **저장소 연결**
   - "Connect a repository" 선택
   - GitHub 계정 연결
   - 배포할 저장소 선택

3. **자동 설정**
   - Render가 `render.yaml`을 자동 감지
   - 설정이 자동으로 적용됨

4. **배포 시작**
   - "Create Web Service" 클릭
   - 자동 빌드 및 배포 시작
   - 약 5-10분 후 배포 완료

### 4. 배포 후 확인

배포가 완료되면 Render가 제공하는 URL로 접속:
```
https://your-app-name.onrender.com
```

#### 테스트
```bash
# 헬스 체크
curl https://your-app-name.onrender.com/health

# 파일 업로드
curl -X POST -F "files=@test.txt" https://your-app-name.onrender.com/api/upload

# 파일 다운로드
curl -O https://your-app-name.onrender.com/download/{file_id}
```

## Docker 배포 (선택사항)

### Docker Hub에 이미지 푸시

```bash
# 이미지 빌드
docker build -t your-username/file-share:latest .

# Docker Hub 로그인
docker login

# 이미지 푸시
docker push your-username/file-share:latest
```

### Docker로 실행

```bash
# 이미지 실행
docker run -d -p 9000:9000 --name file-share your-username/file-share:latest

# 로그 확인
docker logs file-share

# 중지
docker stop file-share

# 재시작
docker start file-share
```

## 환경 변수 설정

### Render에서 환경 변수 설정

1. Render 대시보드에서 서비스 선택
2. "Environment" 탭 클릭
3. 환경 변수 추가:
   - `PORT`: 자동으로 설정됨 (기본값: 10000)

### Docker에서 환경 변수 설정

```bash
docker run -d -p 9000:9000 -e PORT=9000 --name file-share your-username/file-share:latest
```

## 업데이트 배포

### Render (자동 배포)

```bash
# 코드 수정 후
git add .
git commit -m "Update: description of changes"
git push origin main

# Render가 자동으로 재배포
```

### Docker (수동 배포)

```bash
# 새 이미지 빌드
docker build -t your-username/file-share:latest .

# 이미지 푸시
docker push your-username/file-share:latest

# 컨테이너 재시작
docker stop file-share
docker rm file-share
docker pull your-username/file-share:latest
docker run -d -p 9000:9000 --name file-share your-username/file-share:latest
```

## 문제 해결

### Render 로그 확인

1. Render 대시보드에서 서비스 선택
2. "Logs" 탭에서 실시간 로그 확인

### 일반적인 문제

#### 1. 포트 바인딩 오류
- Render는 `PORT` 환경 변수를 자동 설정
- 코드에서 `os.Getenv("PORT")`로 포트 읽기 (이미 적용됨)

#### 2. 업로드 디렉토리 권한 오류
- `uploads/` 디렉토리는 자동 생성됨
- 권한 문제 시 코드에서 `os.MkdirAll(uploadDir, 0755)` 확인

#### 3. 파일 업로드 실패
- 파일 크기 제한: 50MB
- 허용된 파일 형식 확인
- MIME 타입 검증 통과 여부 확인

## 모니터링

### Render 모니터링

- Render 대시보드에서 자동으로 제공:
  - CPU 사용량
  - 메모리 사용량
  - 네트워크 트래픽
  - 요청 응답 시간

### 헬스 체크 엔드포인트

```bash
curl https://your-app-name.onrender.com/health
```

응답:
```json
{
  "status": "ok",
  "service": "file-sharing"
}
```

## 보안 권장사항

### 프로덕션 환경

1. **HTTPS 사용** (Render가 자동 제공)
2. **속도 제한 추가** (선택사항)
3. **인증 시스템 추가** (선택사항)
4. **파일 만료 정책 구현** (선택사항)
5. **데이터베이스 연동** (파일 메타데이터 영구 저장)

### 현재 구현된 보안

- ✅ 파일 크기 제한 (50MB)
- ✅ MIME 타입 검증
- ✅ 확장자 검증
- ✅ 경로 탐색 공격 방지
- ✅ 파일명 새니타이징
- ✅ 보안 HTTP 헤더
- ✅ CORS 설정

## 비용

### Render Free Tier

- 무료 플랜 제공
- 월 750시간 실행
- 비활성 시 자동 슬립 (15분 무활동 후)
- 첫 요청 시 웨이크업 (약 30초 소요)

### 유료 플랜

- 항상 실행 상태 유지
- 더 많은 리소스
- 자세한 내용: https://render.com/pricing

## 추가 참고사항

- 업로드된 파일은 서버 재시작 시에도 유지됨
- 파일 메타데이터는 메모리에 저장 (재시작 시 손실)
- 프로덕션에서는 데이터베이스 사용 권장
