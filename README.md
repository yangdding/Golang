# 파일 공유 서비스

Go로 구축된 안전한 웹 기반 파일 업로드 및 공유 서비스입니다.

## 주요 기능

- 드래그 앤 드롭 지원 웹 UI를 통한 파일 업로드
- 자동 공유 링크 생성
- curl/CLI를 위한 완전한 API 지원
- 포괄적인 보안 검증
- 배포 준비 완료 (Render, Docker 등)

## 보안 기능

- 파일 크기 제한 (최대 50MB)
- MIME 타입 검증
- 확장자 검증 (.jpg, .png, .gif, .pdf, .txt, .zip)
- 경로 탐색 공격 방지
- 파일명 새니타이징
- 보안 헤더 (X-Content-Type-Options, X-Frame-Options 등)
- API 접근을 위한 CORS 지원

## 지원 파일 형식

- **이미지**: JPG, PNG, GIF
- **문서**: PDF, TXT
- **압축**: ZIP

## 빠른 시작

### 로컬 개발 환경

1. **Go 설치** (아직 설치하지 않은 경우)
   - 다운로드: https://go.dev/dl/
   - 확인: `go version`

2. **의존성 설치**
   ```bash
   go mod tidy
   ```

3. **서버 실행**
   ```bash
   go run main.go
   ```

4. **서비스 접속**
   - 웹 UI: `http://localhost:9000`
   - 헬스 체크: `http://localhost:9000/health`

## API 사용법 (curl)

### PowerShell 사용자

PowerShell에서 `curl`은 `Invoke-WebRequest`의 별칭입니다. `curl.exe`를 사용하세요:

```powershell
curl.exe -X POST -F "files=@photo.jpg" http://localhost:9000/api/upload
```

### 파일 업로드
```bash
curl -X POST -F "files=@example.jpg" http://localhost:9000/api/upload
```

**응답 예시:**
```json
{
  "message": "Upload successful",
  "files": [
    {
      "id": "abc123...",
      "original_name": "example.jpg",
      "size": 12345,
      "upload_time": "2025-10-24T...",
      "mime_type": "image/jpeg"
    }
  ]
}
```

### 여러 파일 업로드
```bash
curl -X POST -F "files=@file1.jpg" -F "files=@file2.pdf" http://localhost:9000/api/upload
```

### 파일 정보 조회
```bash
curl http://localhost:9000/api/info/{file_id}
```

### 파일 다운로드
```bash
# 원본 파일명으로 다운로드
curl -O -J http://localhost:9000/download/{file_id}

# 사용자 지정 파일명으로 다운로드
curl -o myfile.jpg http://localhost:9000/download/{file_id}
```

## 배포

### Render

1. GitHub에 코드 푸시
2. Render에서 새 Web Service 생성
3. GitHub 저장소 연결
4. Render가 자동으로 `render.yaml`을 감지하고 배포

서비스 접속: `https://your-app.onrender.com`

### Docker

```bash
# 빌드
docker build -t file-share .

# 실행
docker run -p 9000:9000 file-share
```

### 환경 변수

- `PORT`: 서버 포트 (기본값: 9000)

## 프로젝트 구조

```
GOlang/
├── main.go              # 메인 애플리케이션
├── go.mod               # Go 모듈 의존성
├── go.sum               # 의존성 체크섬
├── Dockerfile           # Docker 설정
├── render.yaml          # Render 배포 설정
├── README.md            # 이 파일
└── uploads/             # 업로드된 파일 저장 디렉토리
```

## 보안 검증 세부사항

### 경로 탐색 공격 방지
- 모든 파일명은 `filepath.Base()`를 사용하여 새니타이징됨
- 파일 ID는 32자 16진수 문자열로 검증됨
- 절대 경로를 확인하여 파일이 업로드 디렉토리 내에 있는지 확인

### 파일 검증
1. **크기 검사**: 파일당 최대 50MB
2. **확장자 검사**: 허용된 확장자만 (.jpg, .png, .gif, .pdf, .txt, .zip)
3. **MIME 타입 검사**: Content-Type 헤더 검증
4. **파일명 새니타이징**: 경로 구분자 및 제어 문자 제거

### HTTP 보안 헤더
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`

## 실제 사용 예시

### 웹 UI를 통한 사용
1. 브라우저에서 `http://localhost:9000` 접속
2. 파일을 드래그 앤 드롭하거나 클릭하여 선택
3. "Upload" 버튼 클릭
4. 생성된 공유 링크 복사
5. 링크를 통해 파일 다운로드 또는 공유

### curl API를 통한 사용
```bash
# 1. 파일 업로드
curl -X POST -F "files=@document.pdf" http://localhost:9000/api/upload

# 2. 응답에서 file_id 확인
# {"message":"Upload successful","files":[{"id":"abc123...","original_name":"document.pdf",...}]}

# 3. 공유 링크 생성
# http://localhost:9000/download/abc123...

# 4. 다른 사용자가 다운로드
curl -O http://localhost:9000/download/abc123...
```

## 배포 환경에서 사용

### Render 배포 후
```bash
# 파일 업로드
curl -X POST -F "files=@image.jpg" https://your-app.onrender.com/api/upload

# 파일 다운로드
curl -O https://your-app.onrender.com/download/{file_id}

# 헬스 체크
curl https://your-app.onrender.com/health
```

## 라이선스

MIT License

"# Golang"  
