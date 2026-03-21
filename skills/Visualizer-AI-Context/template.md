# Mẫu Code Khuyến Cáo
Với Drawio, không cần thiết phải dùng đống XML thư viện nặng nề để xây lại cây DOM XML mới rồi build xml file. Ta có ID dạng String, nên dùng Regex gạch chéo ID và replace style rất nhanh.

// Dòng code minh hoạ tốc độ cao
updatedXML := strings.Replace(rawXML, "id=\"abc\" style=\"", "id=\"abc\" style=\"fillColor=#f8cecc;", 1)
